package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/jackc/pgx"
	jsoniter "github.com/json-iterator/go"
	"github.com/urfave/cli"
)

var pool *pgx.ConnPool

func qHandler(w http.ResponseWriter, r *http.Request) {
	sql := r.URL.Query().Get("query")
	w.Header().Set("Content-Type", "application/json")

	if len(sql) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"message\": \"Please provide 'query' query-string param\"}"))
		return
	}

	conn, err := pool.Acquire()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Cannot acquire DB connection\"}"))
		return
	}

	defer pool.Release(conn)

	stream := jsoniter.ConfigFastest.BorrowStream(w)
	defer jsoniter.ConfigFastest.ReturnStream(stream)

	rows, err := conn.Query(sql)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		stream.WriteVal(map[string]string{
			"message": err.Error(),
		})
		stream.Flush()

		return
	}

	defer rows.Close()

	stream.WriteObjectStart()
	stream.WriteObjectField("columns")
	stream.WriteVal(rows.FieldDescriptions())
	stream.WriteMore()
	stream.WriteObjectField("rows")
	stream.WriteArrayStart()

	hasRows := rows.Next()

	for hasRows {
		vals, err := rows.Values()

		if err == nil {
			stream.WriteVal(vals)
		} else {
			stream.WriteNil()
		}

		hasRows = rows.Next()

		if hasRows {
			stream.WriteMore()
		}
	}

	stream.WriteArrayEnd()
	stream.WriteObjectEnd()

	stream.Flush()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	conn, err := pool.Acquire()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Cannot acquire DB connection\"}"))
		return
	}

	defer pool.Release(conn)

	rows, err := conn.Query("SELECT 1+1")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{ \"message\": \"cannot perform query\" }"))
		return
	}

	defer rows.Close()

	w.Write([]byte("{ \"message\": \"ich bin gesund\" }"))
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				logger.Println(r.Method, r.URL, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// WebAction starts HTTP server and serves basic FB API
func WebCommand(c *cli.Context) error {
	webHost := c.String("webhost")
	webPort := c.Uint("webport")
	addr := fmt.Sprintf("%s:%d", webHost, webPort)
	var box http.FileSystem
	if os.Getenv("DEV") == "" {
		box = packr.NewBox("./web")
	} else {
		box = http.Dir("./web")
	}

	mainConfig := GetPgxConnectionConfig(nil)

	connStr := fmt.Sprintf("dbname=%s sslmode=disable user=%s password=%s host=%s port=%d",
		mainConfig.Database, mainConfig.User, mainConfig.Password, mainConfig.Host, mainConfig.Port)

	var err error
	pool, err = pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: mainConfig})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect to database:", err)
		os.Exit(1)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	logger.Printf("Connected to database %s\n", connStr)

	router := http.NewServeMux()
	router.Handle("/", http.FileServer(box))
	router.HandleFunc("/q", qHandler)
	router.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:         addr,
		Handler:      logging(logger)(router),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := server.Shutdown(context.Background()); err != nil {
			logger.Printf("HTTP server Shutdown: %v\n", err)
		}
		close(idleConnsClosed)
	}()

	logger.Printf("Starting web server on %s\n", addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", addr, err)
	}

	logger.Println("Server stopped")
	return nil
}
