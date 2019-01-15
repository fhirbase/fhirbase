package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

type bundleFile struct {
	file *os.File
	gzr  *gzip.Reader
}

func openFile(fileName string) (*bundleFile, error) {
	result := new(bundleFile)

	f, err := os.OpenFile(fileName, os.O_RDONLY, 0644)

	if err != nil {
		return nil, errors.Wrap(err, "cannot open bundle file")
	}

	result.file = f

	gzr, err := gzip.NewReader(result.file)

	if err != nil {
		result.file.Seek(0, 0)
		result.gzr = nil
	} else {
		result.gzr = gzr
	}

	return result, nil
}

func (bf *bundleFile) Read(p []byte) (n int, err error) {
	if bf.gzr != nil {
		return bf.gzr.Read(p)
	}

	return bf.file.Read(p)
}

func (bf *bundleFile) Close() {
	defer bf.file.Close()

	if bf.gzr != nil {
		bf.gzr.Close()
	}
}

func (bf *bundleFile) Rewind() {
	bf.file.Seek(0, 0)

	if bf.gzr != nil {
		bf.gzr.Close()
		bf.gzr.Reset(bf.file)
	}
}

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

type singleResourceBundle struct {
	file        *bundleFile
	alreadyRead bool
}

func newSingleResourceBundle(f *bundleFile) (*singleResourceBundle, error) {
	b := new(singleResourceBundle)

	b.file = f
	b.alreadyRead = false

	return b, nil
}

func (b *singleResourceBundle) Close() {
	b.file.Close()
}

func (b *singleResourceBundle) Count() int {
	return 1
}

func (b *singleResourceBundle) Next() (map[string]interface{}, error) {
	if b.alreadyRead {
		return nil, io.EOF
	}

	content, err := ioutil.ReadAll(b.file)

	if err != nil {
		return nil, errors.Wrap(err, "cannot read file content")
	}

	iter := jsoniter.ConfigFastest.BorrowIterator(content)
	defer jsoniter.ConfigFastest.ReturnIterator(iter)

	res := iter.Read()

	if res == nil {
		return nil, errors.Wrap(iter.Error, "cannot read resource from file")
	}

	resMap, ok := res.(map[string]interface{})

	if !ok {
		return nil, fmt.Errorf("got non-object value in the entries array")
	}

	b.alreadyRead = true

	return resMap, nil
}

type ndjsonBundle struct {
	count   int
	file    *bundleFile
	reader  *bufio.Reader
	curline int
}

type fhirBundle struct {
	count   int
	file    *bundleFile
	curline int
	iter    *jsoniter.Iterator
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
		fmt.Printf("%s: got non-object value in the entries array, skipping rest of the file\n", b.file.file.Name())
		return nil, io.EOF
	}

	res, ok := entryMap["resource"]

	if !ok {
		fmt.Printf("%s: cannot get entry.resource attribute, skipping rest of the file\n", b.file.file.Name())
		return nil, io.EOF
	}

	resMap, ok := res.(map[string]interface{})

	if !ok {
		fmt.Printf("%s: got non-object value at entry.resource, skipping rest of the file\n", b.file.file.Name())
		return nil, io.EOF
	}

	return resMap, nil
}

func newFhirBundle(f *bundleFile) (*fhirBundle, error) {
	var result fhirBundle

	result.file = f
	result.iter = jsoniter.Parse(jsoniter.ConfigFastest, result.file, 32*1024)

	err := goToEntriesInFhirBundle(result.iter)

	if err != nil {
		return nil, errors.Wrap(err, "cannot find `entry` key in the bundle")
	}

	linesCount, err := countEntriesInBundle(result.iter)

	result.file.Rewind()
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

func (b *ndjsonBundle) Close() {
	b.file.Close()
}

func (b *ndjsonBundle) Count() int {
	return b.count
}

func (b *ndjsonBundle) Next() (map[string]interface{}, error) {
	line, err := b.reader.ReadBytes('\n')

	iter := jsoniter.ConfigFastest.BorrowIterator(line)
	defer jsoniter.ConfigFastest.ReturnIterator(iter)

	if err != nil {
		return nil, err
	}

	if iter.WhatIsNext() != jsoniter.ObjectValue {
		fmt.Printf("%s: Expecting to get JSON object at the root of the resource, got `%s` at line %d, skipping rest of the file\n", b.file.file.Name(), strings.Trim(string(line), "\n"), b.curline)
		return nil, io.EOF
	}

	b.curline++

	result := iter.Read()

	return result.(map[string]interface{}), iter.Error
}

func newNdjsonBundle(f *bundleFile) (*ndjsonBundle, error) {
	var result ndjsonBundle
	result.file = f
	result.reader = bufio.NewReader(result.file)

	linesCount, err := countLinesInReader(result.reader)

	if err != nil {
		return nil, errors.Wrap(err, "cannot count lines in ndjson bundle")
	}

	result.file.Rewind()

	result.count = linesCount

	return &result, nil
}

type multifileBundle struct {
	count          int
	bundles        []bundle
	currentBndlIdx int
}

func newMultifileBundle(fileNames []string) (*multifileBundle, error) {
	var result multifileBundle
	result.bundles = make([]bundle, 0, len(fileNames))
	result.count = 0
	result.currentBndlIdx = 0

	for _, fileName := range fileNames {
		f, err := openFile(fileName)

		if err != nil {
			fmt.Printf("Cannot open %s: %v\n", fileName, err)
			continue
		}

		bndlType, err := guessBundleType(f)

		if err != nil {
			fmt.Printf("Cannot determine type of %s: %v\n", fileName, err)
			f.Close()
			continue
		}

		f.Rewind()

		var bndl bundle

		if bndlType == ndjsonBundleType {
			bndl, err = newNdjsonBundle(f)
		} else if bndlType == fhirBundleType {
			bndl, err = newFhirBundle(f)
		} else if bndlType == singleResourceBundleType {
			bndl, err = newSingleResourceBundle(f)
		} else {
			fmt.Printf("cannot create bundle for %s\n", fileName)
			continue
		}

		if err != nil {
			fmt.Printf("%s: cannot create bundle\n%e\n", f.file.Name(), err)
			defer f.Close()
			bndl = nil
		}

		if bndl != nil {
			result.bundles = append(result.bundles, bndl)
			result.count = result.count + bndl.Count()
		}
	}

	return &result, nil
}

func (b *multifileBundle) Count() int {
	return b.count
}

func (b *multifileBundle) Close() {
	for _, bndl := range b.bundles {
		if bndl != nil {
			b.Close()
		}
	}

	b.currentBndlIdx = -1
}

func (b *multifileBundle) Next() (map[string]interface{}, error) {
	if b.currentBndlIdx >= len(b.bundles) {
		return nil, io.EOF
	}

	currentBndl := b.bundles[b.currentBndlIdx]

	// if b.currentBndl == nil {
	// 	b.currentBndlIdx = b.currentBndlIdx + 1

	// 	if b.currentBndlIdx > len(b.fileNames)-1 {
	// 		return nil, io.EOF
	// 	}

	// 	currentBndl, err := newFhirBundle(b.fileNames[b.currentBndlIdx])

	// 	if err != nil {
	// 		b.currentBndlIdx = b.currentBndlIdx + 1
	// 		return nil, errors.Wrap(err, "cannot create bundle")
	// 	}

	// 	b.currentBndl = currentBndl
	// }

	res, err := currentBndl.Next()

	if err != nil {
		if err == io.EOF {
			currentBndl.Close()
			b.bundles[b.currentBndlIdx] = nil
			b.currentBndlIdx = b.currentBndlIdx + 1

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

	if batch != nil {
		batch.Send(context.Background(), nil)
		batch.Close()
	}

	return nil
}

func prewalkDirs(fileNames []string) ([]string, error) {
	result := make([]string, 0)

	for _, fn := range fileNames {
		fi, err := os.Stat(fn)

		switch {
		case err != nil:
			return nil, err
		case fi.IsDir():
			err = filepath.Walk(fn, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					result = append(result, path)
				}

				return err
			})

			if err != nil {
				return nil, err
			}
		default:
			result = append(result, fn)
		}
	}

	return result, nil
}

func loadFiles(files []string, ldr loader, memUsage bool) error {
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

	loadDuration := int(time.Since(startTime).Seconds())

	submitLoadEvent(insertedCounts, loadDuration)

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

	var bulkLoad bool

	if strings.HasPrefix(c.Args().Get(0), "http") {
		bulkLoad = true
	} else {
		bulkLoad = false
	}

	fhirVersion := c.GlobalString("fhir")
	mode := c.String("mode")
	var ldr loader

	if bulkLoad && !c.IsSet("mode") {
		mode = "copy"
	}

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

	if bulkLoad {
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

		return loadFiles(files, ldr, memUsage)
	}

	files, err := prewalkDirs(c.Args())

	if err != nil {
		return errors.Wrap(err, "cannot prewalk directories")
	}

	return loadFiles(files, ldr, memUsage)
}
