package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gobuffalo/packr"
	"github.com/jackc/pgx"
	jsoniter "github.com/json-iterator/go"
	"github.com/urfave/cli"
)

var pool *pgx.ConnPool

// WebAction starts HTTP server and serves basic FB API
func WebCommand(c *cli.Context) error {
	webHost := c.String("webhost")
	webPort := c.Uint("webport")
	addr := fmt.Sprintf("%s:%d", webHost, webPort)
	box := packr.NewBox("./web")

	mainConfig := GetPgxConnectionConfig(nil)

	connStr := fmt.Sprintf("dbname=%s sslmode=disable user=%s password=%s host=%s port=%d",
		mainConfig.Database, mainConfig.User, mainConfig.Password, mainConfig.Host, mainConfig.Port)

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: mainConfig})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect to database:", err)
		os.Exit(1)
	}

	fmt.Printf("Connected to database %s\n", connStr)

	http.Handle("/", http.FileServer(box))

	http.HandleFunc("/q", func(w http.ResponseWriter, r *http.Request) {
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
	})

	fmt.Printf("Starting web server on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}
