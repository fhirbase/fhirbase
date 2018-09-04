package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"text/tabwriter"

	"compress/gzip"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	jsoniter "github.com/json-iterator/go"
	"github.com/urfave/cli"
	pb "gopkg.in/cheggaaa/pb.v2"
)

const BatchSize = 2000

type bundle interface {
	Next() (map[string]interface{}, error)
	Close()
	Count() int
}

type multilineBundle struct {
	count   int
	file    *os.File
	gzr     *gzip.Reader
	reader  *bufio.Reader
	curline int
}

func (b *multilineBundle) Close() {
	if b.gzr != nil {
		b.gzr.Close()
	}

	defer b.file.Close()
}

func (b *multilineBundle) Count() int {
	return b.count
}

func (b *multilineBundle) Next() (map[string]interface{}, error) {
	line, err := b.reader.ReadBytes('\n')

	iter := jsoniter.ConfigDefault.BorrowIterator(line)
	defer jsoniter.ConfigDefault.ReturnIterator(iter)

	if err != nil {
		return nil, err
	}

	if iter.WhatIsNext() != jsoniter.ObjectValue {
		return nil, fmt.Errorf("Expecting to get JSON object at the root of the resource, got `%s` at line %d", strings.Trim(string(line), "\n"), b.curline)
	}

	b.curline++

	result := iter.Read()

	return result.(map[string]interface{}), iter.Error
}

func newMultilineBundle(fileName string) (*multilineBundle, error) {
	var result multilineBundle

	file, err := os.Open(fileName)

	if err != nil {
		return nil, err
	}

	result.file = file

	zr, err := gzip.NewReader(file)

	if err == gzip.ErrHeader {
		file.Seek(0, 0)
		result.gzr = nil
		result.reader = bufio.NewReader(result.file)
	} else {
		result.gzr = zr
		result.reader = bufio.NewReader(zr)
	}

	linesCount, err := countLinesInReader(result.reader)
	result.file.Seek(0, 0)

	if err != nil {
		return nil, err
	}

	if result.gzr != nil {
		result.gzr.Close()
		result.gzr.Reset(result.file)
	}

	result.count = linesCount

	return &result, nil
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func countLinesInReader(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

func performLoad(db *pgx.Conn, bndl bundle, batchSize uint, progressCb func(cur uint, curType string, total uint)) error {
	tx, _ := db.Begin()
	batch := tx.BeginBatch()
	curResource := uint(0)
	totalCount := uint(bndl.Count())
	var err error

	for err == nil {
		var resource map[string]interface{}
		resource, err = bndl.Next()

		if err == nil {
			resourceType, _ := resource["resourceType"].(string)

			batch.Queue(fmt.Sprintf("INSERT INTO %s (id, txid, status, resource) VALUES (gen_random_uuid(), 0, 'created', $1)", strings.ToLower(resourceType)), []interface{}{resource}, []pgtype.OID{pgtype.JSONBOID}, nil)

			if curResource%batchSize == 0 || curResource == totalCount-1 {
				// PrintMemUsage()
				batch.Send(context.Background(), nil)
				batch.Close()

				if curResource != totalCount-1 {
					batch = db.BeginBatch()
				} else {
					batch = nil
				}
			}

			curResource++
			progressCb(curResource, resourceType, totalCount)
		} else {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()

	return nil
}

// LoadCommand loads FHIR schema into database
func LoadCommand(c *cli.Context) error {
	db := GetConnection(nil)
	defer db.Close()

	batchSize := c.Uint("batchsize")

	bndl, err := newMultilineBundle(c.Args()[0])
	defer bndl.Close()

	if err != nil {
		return err
	}

	insertedCounts := make(map[string]uint)
	bar := pb.Full.Start(bndl.Count())
	bar.SetWidth(100)

	err = performLoad(db, bndl, batchSize, func(cur uint, curType string, total uint) {
		insertedCounts[curType] = insertedCounts[curType] + 1
		bar.Increment()
	})

	bar.Finish()
	fmt.Printf("Done, inserted %d resources:\n", bndl.Count())

	tblw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)

	for rt, cnt := range insertedCounts {
		fmt.Fprintf(tblw, "%s\t %d\n", rt, cnt)
	}

	tblw.Flush()

	return nil
}
