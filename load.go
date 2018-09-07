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
	"time"

	"compress/gzip"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	jsoniter "github.com/json-iterator/go"
	"github.com/urfave/cli"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

type bundle interface {
	Next() (map[string]interface{}, error)
	Close()
	Count() int
}

type multilineBundle struct {
	count    int
	fileName string
	file     *os.File
	gzr      *gzip.Reader
	reader   *bufio.Reader
	curline  int
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
	result.fileName = fileName

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

type multifileBundle struct {
	count          int
	fileNames      []string
	currentBndlIdx int
	currentBndl    bundle
}

func newMultifileBundle(fileNames []string) (*multifileBundle, error) {
	var result multifileBundle
	result.fileNames = fileNames
	result.count = 0
	result.currentBndlIdx = -1

	for _, fileName := range result.fileNames {
		bndl, err := newMultilineBundle(fileName)

		if err != nil {
			return nil, err
		}

		result.count = result.count + bndl.Count()
		bndl.Close()
	}

	return &result, nil
}

func (b *multifileBundle) Count() int {
	return b.count
}

func (b *multifileBundle) Close() {
	if b.currentBndl != nil {
		b.currentBndl.Close()
		b.currentBndl = nil
		b.currentBndlIdx = -1
	}
}

func (b *multifileBundle) Next() (map[string]interface{}, error) {
	if b.currentBndl == nil {
		b.currentBndlIdx = b.currentBndlIdx + 1

		if b.currentBndlIdx > len(b.fileNames)-1 {
			return nil, io.EOF
		}

		currentBndl, err := newMultilineBundle(b.fileNames[b.currentBndlIdx])

		if err != nil {
			b.currentBndlIdx = b.currentBndlIdx + 1
			return nil, err
		}

		b.currentBndl = currentBndl
	}

	res, err := b.currentBndl.Next()

	if err != nil {
		if err == io.EOF {
			b.currentBndl.Close()
			b.currentBndl = nil
			return b.Next()
		}
		return nil, err
	}

	return res, nil
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

func performLoad(db *pgx.Conn, bndl bundle, batchSize uint, progressCb func(cur uint, curType string, total uint, duration time.Duration)) error {
	tx, _ := db.Begin()
	batch := tx.BeginBatch()
	curResource := uint(0)
	totalCount := uint(bndl.Count())
	var err error

	for err == nil {
		startTime := time.Now()
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
			progressCb(curResource, resourceType, totalCount, time.Since(startTime))
		} else {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()

	return nil
}

func loadFiles(files []string, batchSize uint) error {
	db := GetConnection(nil)
	defer db.Close()

	bndl, err := newMultifileBundle(files)
	defer bndl.Close()

	if err != nil {
		return err
	}

	insertedCounts := make(map[string]uint)
	bars := mpb.New(
		mpb.WithWidth(100),
	)

	bar := bars.AddBar(int64(bndl.Count()),
		mpb.AppendDecorators(
			decor.Percentage(decor.WC{W: 3}),
			decor.AverageETA(decor.ET_STYLE_MMSS, decor.WC{W: 6}),
		),
		mpb.PrependDecorators(decor.CountersNoUnit("%d / %d", decor.WC{W: 10})))

	err = performLoad(db, bndl, batchSize, func(cur uint, curType string, total uint, duration time.Duration) {
		insertedCounts[curType] = insertedCounts[curType] + 1
		bar.IncrBy(1, duration)
	})

	bars.Wait()

	if err != nil && err != io.EOF {
		return err
	}

	fmt.Printf("Done, inserted %d resources:\n", bndl.Count())

	tblw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)

	for rt, cnt := range insertedCounts {
		fmt.Fprintf(tblw, "%s\t %d\n", rt, cnt)
	}

	tblw.Flush()

	return nil
}

// LoadCommand loads FHIR schema into database
func LoadCommand(c *cli.Context) error {
	if c.NArg() == 0 {
		cli.ShowCommandHelpAndExit(c, "load", 1)
		return nil
	}

	if strings.HasPrefix(c.Args().Get(0), "http") {
		fileHndlrs, err := getBulkData(c.Args().Get(0))

		if err != nil {
			return err
		}

		files := make([]string, 0, len(fileHndlrs))

		defer func() {
			for _, fn := range files {
				os.Remove(fn)
			}
		}()

		for _, f := range fileHndlrs {
			files = append(files, f.Name())
			f.Close()
		}

		return loadFiles(files, c.Uint("batchsize"))
	}

	return loadFiles(c.Args(), c.Uint("batchsize"))
}
