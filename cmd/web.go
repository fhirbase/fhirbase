package cmd

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	db "github.com/fhirbase/fhirbase/db"
	"github.com/jackc/pgx/v5/pgxpool"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Starts web server with primitive UI to perform SQL queries from the browser",
	Long: `
Starts a simple web server to invoke SQL queries from the browser UI.

You can specify web server's host and port with "--webhost" and
"--webport" flags. If "--webhost" flag is empty (set to blank string)
then web server will listen on all available network interfaces.`,
	Example: "fhirbase [--fhir=FHIR version] web",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetString("webhost") == "" {
			viper.Set("webhost", "localhost")
		}
		fmt.Println("Web server host: ", viper.GetString("webhost"))
		fmt.Println("Web server port: ", viper.GetUint("webport"))
		//
		// Do Stuff Here
		ctx := cmd.Context()
		WebCommand(ctx)

	},
}
var webhost string
var webport uint

func init() {
	rootCmd.AddCommand(webCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	webCmd.PersistentFlags().String("foo", "", "A help for foo")
	webCmd.PersistentFlags().StringVarP(&webhost, "webhost", "", "localhost", "Web server host")
	webCmd.PersistentFlags().UintVarP(&webport, "webport", "", 3000, "Web server port")
	viper.BindPFlag("webhost", webCmd.PersistentFlags().Lookup("webhost"))
	viper.BindPFlag("webport", webCmd.PersistentFlags().Lookup("webport"))
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// webCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func qHandler(ctx context.Context, db *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	sql := r.URL.Query().Get("query")
	w.Header().Set("Content-Type", "application/json")

	if len(sql) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"message\": \"Please provide 'query' query-string param\"}"))
		return
	}

	conn, err := db.Acquire(ctx)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Cannot acquire DB connection\"}"))
		return
	}

	defer conn.Release()

	stream := jsoniter.ConfigFastest.BorrowStream(w)
	defer jsoniter.ConfigFastest.ReturnStream(stream)

	rows, err := conn.Query(context.Background(), sql)

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

func healthHandler(ctx context.Context, db *pgxpool.Pool, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	conn, err := db.Acquire(ctx)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Cannot acquire DB connection\"}"))
		return
	}

	defer conn.Release()

	rows, err := conn.Query(context.Background(), "SELECT 1+1")

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

//go:embed web/*
var webFiles embed.FS

// WebAction starts HTTP server and serves basic FB API
func WebCommand(ctx context.Context) error {

	webHost := viper.GetString("webhost")
	webPort := viper.GetUint("webport")
	addr := fmt.Sprintf("%s:%d", webHost, webPort)

	database, err := db.GetConnection()

	if err != nil {
		return fmt.Errorf("Error acquiring connection: %v", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	logger.Printf("Connected to database %s\n", database.Config().ConnString())

	router := http.NewServeMux()
	webFS, err := fs.Sub(webFiles, "web")
	router.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(webFS))))
	router.HandleFunc("/q", func(w http.ResponseWriter, r *http.Request) {

		qHandler(ctx, database, w, r)

	})

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {

		healthHandler(ctx, database, w)

	})

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
