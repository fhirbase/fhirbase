package main

import (
	"bufio"
	"compress/gzip"
	"database/sql"
	"log"
	"os"
	"reflect"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/lib/pq"
)

func BytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{bh.Data, bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

func main() {
	db, err := sql.Open("postgres", "dbname=fbloader sslmode=disable user=postgres password=postgres")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("connected to psql")

	file, err := os.Open("./sample-data.gzip")

	if err != nil {
		log.Fatal(err)
	}

	zr, err := gzip.NewReader(file)

	if err != nil {
		log.Fatal(err)
	}

	bufr := bufio.NewReader(zr)

	if err != nil {
		log.Fatal(err)
	}

	cp := pq.CopyIn("resources", "resource")

	txn, err := db.Begin()
	if err != nil {
		log.Fatalf("to.Begin: %v", err)
	}

	stmt, err := txn.Prepare(cp)
	if err != nil {
		log.Fatalf("txn.Prepare: %v", err)
	}

	for err == nil {
		var line string
		line, err = bufr.ReadString('\n')

		iter := jsoniter.ConfigFastest.BorrowIterator([]byte(line))
		defer jsoniter.ConfigFastest.ReturnIterator(iter)

		stream := jsoniter.ConfigFastest.BorrowStream(nil)
		defer jsoniter.ConfigFastest.ReturnStream(stream)

		// force parsing
		var res interface{} = iter.Read()
		// log.Printf(obj["resourcxeType"].(string))

		stream.WriteVal(res)

		if res != nil {
			_, err = stmt.Exec(BytesToString(stream.Buffer()))

			if err != nil {
				log.Printf("[copy] ERROR (stmt.Exec): %s", err)
				log.Printf("%s", line)
			}

		}

	}

	stmt.Exec()
	stmt.Close()
	txn.Commit()
}
