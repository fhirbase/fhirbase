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
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

type bundleType int

const (
	ndjsonBundleType bundleType = iota
	fhirBundleType
	singleResourceBundleType
	unknownBundleType
)

type bundle interface {
	Next() (map[string]interface{}, error)
	Close()
	Count() int
}

type loaderCb func(curType string, duration time.Duration)

type loader interface {
	Load(db *pgx.Conn, bndl bundle, cb loaderCb) error
}

type copyFromBundleSource struct {
	bndl        bundle
	err         error
	res         map[string]interface{}
	cb          loaderCb
	currentRt   string
	prevTime    time.Time
	fhirVersion string
}

func isCompleteJSONObject(s string) bool {
	numBraces := 0
	inString := false
	escaped := false

	for _, b := range s {
		if !escaped {
			if !inString {
				if b == '{' {
					numBraces = numBraces + 1
				} else if b == '}' {
					numBraces = numBraces - 1
				} else if b == '"' {
					inString = true
				}
			} else {
				if b == '"' {
					inString = false
				} else if b == '\\' {
					escaped = true
				}
			}
		} else {
			escaped = false
		}
	}

	return numBraces == 0
}

func guessJSONBundleType(r io.Reader) (bundleType, error) {
	iter := jsoniter.Parse(jsoniter.ConfigFastest, r, 32*1024)

	if iter.WhatIsNext() != jsoniter.ObjectValue {
		return unknownBundleType, fmt.Errorf("Expecting to get JSON object at the root of the resource")
	}

	for k := iter.ReadObject(); k != ""; k = iter.ReadObject() {
		if k == "resourceType" {
			rt := iter.ReadString()

			if rt == "Bundle" {
				return fhirBundleType, nil
			} else if rt != "" {
				return singleResourceBundleType, nil
			}

			return unknownBundleType, nil
		}

		iter.Skip()
	}

	return fhirBundleType, nil
}

func guessBundleType(f io.Reader) (bundleType, error) {
	rdr := bufio.NewReader(f)
	firstLine, err := rdr.ReadString('\n')

	if err != nil {
		if err == io.EOF {
			// only one line is available
			return guessJSONBundleType(strings.NewReader(firstLine))
		}

		return unknownBundleType, err
	}

	secondLine, err := rdr.ReadString('\n')

	if err != nil && err != io.EOF {
		return unknownBundleType, err
	}

	if isCompleteJSONObject(firstLine) && isCompleteJSONObject(secondLine) {
		return ndjsonBundleType, nil
	}

	return guessJSONBundleType(io.MultiReader(strings.NewReader(firstLine),
		strings.NewReader(secondLine), rdr))
}

func newCopyFromBundleSource(bndl bundle, fhirVersion string, cb loaderCb) *copyFromBundleSource {
	s := new(copyFromBundleSource)

	s.bndl = bndl
	s.err = nil
	s.cb = cb

	res, _ := bndl.Next()
	rt, _ := res["resourceType"].(string)

	s.res = res
	s.currentRt = rt
	s.prevTime = time.Now()
	s.fhirVersion = fhirVersion

	return s
}

func (s *copyFromBundleSource) Next() bool {
	if s.res != nil {
		return true
	}

	res, err := s.bndl.Next()

	if err != nil {
		s.res = nil

		if err != io.EOF {
			s.err = err
		} else {
			s.currentRt = ""
			s.err = nil
		}

		return false
	}

	nextResourceType, _ := res["resourceType"].(string)

	if nextResourceType != s.currentRt {
		s.currentRt = nextResourceType
		s.res = res
		s.prevTime = time.Now()
		s.err = nil

		return false
	}

	s.res = res
	s.err = nil

	return true
}

func (s *copyFromBundleSource) ResourceType() string {
	return s.currentRt
}

func (s *copyFromBundleSource) Values() ([]interface{}, error) {
	if s.res != nil {
		res := s.res
		s.res = nil

		res, err := doTransform(res, s.fhirVersion)

		if err != nil {
			return nil, errors.Wrap(err, "cannot perform transform")
		}

		id, ok := res["id"].(string)

		if !ok {
			id = uuid.NewV4().String()
		}

		d := time.Since(s.prevTime)
		s.prevTime = time.Now()

		s.cb(s.currentRt, d)

		return []interface{}{id, 0, "created", res}, nil
	}

	return nil, fmt.Errorf("No resource in the source")
}

func (s *copyFromBundleSource) Err() error {
	return s.err
}

type multilineBundle struct {
	count    int
	fileName string
	file     *os.File
	gzr      *gzip.Reader
	reader   *bufio.Reader
	curline  int
}

type fhirBundle struct {
	count    int
	fileName string
	file     *os.File
	curline  int
	iter     *jsoniter.Iterator
}

func (b *fhirBundle) Close() {
	b.file.Close()
}

func (b *fhirBundle) Count() int {
	return b.count
}

func (b *fhirBundle) Next() (map[string]interface{}, error) {
	if !b.iter.ReadArray() {
		return nil, io.EOF
	}

	entry := b.iter.Read()

	if entry == nil {
		return nil, b.iter.Error
	}

	entryMap, ok := entry.(map[string]interface{})

	if !ok {
		return nil, fmt.Errorf("got non-object value in the entries array")
	}

	res, ok := entryMap["resource"]

	if !ok {
		return nil, fmt.Errorf("cannot get entry.resource attribute")
	}

	resMap, ok := res.(map[string]interface{})

	if !ok {
		return nil, fmt.Errorf("got non-object value at entry.resource")
	}

	fmt.Printf("%v\n\n", resMap)

	return resMap, nil
}

func newFhirBundle(fileName string) (*fhirBundle, error) {
	var result fhirBundle
	result.fileName = fileName

	file, err := os.Open(fileName)

	if err != nil {
		return nil, err
	}

	result.file = file
	result.iter = jsoniter.Parse(jsoniter.ConfigFastest, result.file, 32*1024)

	err = goToEntriesInFhirBundle(result.iter)

	if err != nil {
		return nil, errors.Wrap(err, "cannot find `entry` key in the bundle")
	}

	linesCount, err := countEntriesInBundle(result.iter)

	result.file.Seek(0, 0)
	result.iter.Reset(result.file)

	if err != nil {
		return nil, errors.Wrap(err, "cannot reset fhir bundle iterator")
	}

	err = goToEntriesInFhirBundle(result.iter)

	if err != nil {
		return nil, errors.Wrap(err, "cannot find `entry` key in the bundle")
	}

	result.count = linesCount

	return &result, nil
}

func (b *multilineBundle) Close() {
	defer b.file.Close()

	if b.gzr != nil {
		b.gzr.Close()
	}
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
		bndl, err := newFhirBundle(fileName)

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

		currentBndl, err := newFhirBundle(b.fileNames[b.currentBndlIdx])

		if err != nil {
			b.currentBndlIdx = b.currentBndlIdx + 1
			return nil, errors.Wrap(err, "cannot create bundle")
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
		return nil, errors.Wrap(err, "cannot read next entry from bundle")
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
	fmt.Printf("\tNumGC = %v\n\n", m.NumGC)
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

func goToEntriesInFhirBundle(iter *jsoniter.Iterator) error {
	if iter.WhatIsNext() != jsoniter.ObjectValue {
		return fmt.Errorf("Expecting to get JSON object at the root of the FHIR Bundle")
	}

	curAttr := iter.ReadObject()

	for curAttr != "" {
		if curAttr == "entry" && iter.WhatIsNext() == jsoniter.ArrayValue {
			return nil
		}

		iter.Skip()

		curAttr = iter.ReadObject()
	}

	return io.EOF
}

func countEntriesInBundle(iter *jsoniter.Iterator) (int, error) {
	count := 0

	for iter.ReadArray() {
		count = count + 1
		iter.Skip()
	}

	return count, nil
}

type copyLoader struct {
	fhirVersion string
}

type insertLoader struct {
	fhirVersion string
}

func (l *copyLoader) Load(db *pgx.Conn, bndl bundle, cb loaderCb) error {
	src := newCopyFromBundleSource(bndl, l.fhirVersion, cb)

	for src.ResourceType() != "" {
		tableName := strings.ToLower(src.ResourceType())

		_, err := db.CopyFrom(pgx.Identifier{tableName}, []string{"id", "txid", "status", "resource"}, src)

		if err != nil {
			return errors.Wrap(err, "cannot perform COPY command")
		}
	}

	return nil
}

func (l *insertLoader) Load(db *pgx.Conn, bndl bundle, cb loaderCb) error {
	batch := db.BeginBatch()
	curResource := uint(0)
	totalCount := uint(bndl.Count())
	batchSize := uint(2000)
	var err error

	for err == nil {
		startTime := time.Now()
		var resource map[string]interface{}
		resource, err = bndl.Next()

		if err == nil {
			transformedResource, err := doTransform(resource, l.fhirVersion)

			if err != nil {
				fmt.Printf("Error during FB transform: %v\n", err)
			}

			resourceType, _ := resource["resourceType"].(string)
			tblName := strings.ToLower(resourceType)
			id, ok := resource["id"].(string)

			if !ok || id == "" {
				batch.Queue(fmt.Sprintf("INSERT INTO %s (id, txid, status, resource) VALUES (gen_random_uuid()::text, 0, 'created', $1) ON CONFLICT (id) DO NOTHING", tblName), []interface{}{transformedResource}, []pgtype.OID{pgtype.JSONBOID}, nil)
			} else {
				batch.Queue(fmt.Sprintf("INSERT INTO %s (id, txid, status, resource) VALUES ($1, 0, 'created', $2) ON CONFLICT (id) DO NOTHING", tblName), []interface{}{id, transformedResource}, []pgtype.OID{pgtype.TextOID, pgtype.JSONBOID}, nil)
			}

			if curResource%batchSize == 0 || curResource == totalCount-1 {
				batch.Send(context.Background(), nil)
				batch.Close()

				if curResource != totalCount-1 {
					batch = db.BeginBatch()
				} else {
					batch = nil
				}
			}

			curResource++
			cb(resourceType, time.Since(startTime))
		} else {
			return err
		}
	}

	return nil
}

func loadNdjsonFiles(files []string, ldr loader, memUsage bool) error {
	db := GetConnection(nil)
	defer db.Close()

	startTime := time.Now()
	bndl, err := newMultifileBundle(files)

	if err != nil {
		return err
	}

	totalCount := bndl.Count()

	insertedCounts := make(map[string]uint)
	currentIdx := 0

	bars := mpb.New(
		mpb.WithWidth(100),
	)

	bar := bars.AddBar(int64(totalCount),
		mpb.AppendDecorators(
			decor.Percentage(decor.WC{W: 3}),
			decor.AverageETA(decor.ET_STYLE_MMSS, decor.WC{W: 6}),
		),
		mpb.PrependDecorators(decor.CountersNoUnit("%d / %d", decor.WC{W: 10})))

	err = ldr.Load(db, bndl, func(curType string, duration time.Duration) {
		if memUsage && currentIdx%3000 == 0 {
			PrintMemUsage()
		}

		currentIdx = currentIdx + 1
		insertedCounts[curType] = insertedCounts[curType] + 1
		bar.IncrBy(1, duration)
	})

	if err != nil && err != io.EOF {
		bars.Abort(bar, false)
		return err
	}

	bars.Wait()

	loadDuration := time.Since(startTime) / time.Second

	fmt.Printf("Done, inserted %d resources in %d seconds:\n", totalCount, loadDuration)

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

	fhirVersion := c.GlobalString("fhir")
	mode := c.String("mode")
	var ldr loader

	if mode != "copy" && mode != "insert" {
		return fmt.Errorf("invalid value for --mode flag. Possible values are either 'copy' or 'insert'")
	}

	if mode == "copy" {
		ldr = &copyLoader{
			fhirVersion: fhirVersion,
		}
	} else {
		ldr = &insertLoader{
			fhirVersion: fhirVersion,
		}
	}

	memUsage := c.Bool("memusage")

	if strings.HasPrefix(c.Args().Get(0), "http") {
		numWorkers := c.Uint("numdl")
		acceptHdr := c.String("accept-header")
		fileHndlrs, err := getBulkData(c.Args().Get(0), numWorkers, acceptHdr)

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

		return loadNdjsonFiles(files, ldr, memUsage)
	}

	return loadNdjsonFiles(c.Args(), ldr, memUsage)
}
