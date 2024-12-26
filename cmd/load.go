/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"compress/gzip"
	"os"
	"time"

	db "github.com/fhirbase/fhirbase/db"
	jsoniter "github.com/json-iterator/go"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type bundleType int

type bundleFile struct {
	file *os.File
	gzr  *gzip.Reader
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
	Load(ctx context.Context, db *pgxpool.Pool, bndl bundle, cb loaderCb) error
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

type singleResourceBundle struct {
	file        *bundleFile
	alreadyRead bool
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
type copyLoader struct {
	fhirVersion string
}

type insertLoader struct {
	fhirVersion string
}

type multifileBundle struct {
	count          int
	bundles        []bundle
	currentBndlIdx int
}

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:       "load",
	Short:     "A brief description of your command",
	ValidArgs: []string{"filename"},
	Long: `
Load command loads FHIR resources from named source(s) into the
Fhirbase database.

You can provide either single Bulk Data URL or several file paths as
an input.

Fhirbase can read from following file types:

  * NDJSON files
  * transaction or collection FHIR Bundles
  * regular JSON files containing single FHIR resource

Also Fhirbase can read gziped files, so all of the supported file
formats can be additionally gziped.

You are allowed to mix different file formats and gziped/non-gziped
files in a single command input, i.e.:

  fhirbase load *.ndjson.gzip patient-john-doe.json my-tx-bundle.json

Fhirbase automatically detects gzip compression and format of the
input file, so you don't have to provide any additional hints. Even
file name extensions can be ommited, because Fhirbase analyzes file
content, not the file name.

If Bulk Data URL was provided, Fhirbase will download NDJSON files
first (see the help for "bulkget" command) and then load them as a
regular local files. Load command accepts all the command-line flags
accepted by bulkget command.

Fhirbase reads input files sequentially, reading single resource at a
time. And because of PostgreSQL traits it's important if Fhirbase gets
a long enough series of resources of the same type from the provided
input, or it gets resource of a different type on every next read. We
will call those two types of inputs "grouped" and "non-grouped",
respectively. Good example of grouped input is NDJSON files produced
by Bulk Data API server. A set of JSON files from FHIR distribution's
"examples" directory is an example of non-grouped input. Because
Fhirbase reads resources one by one and do not load the whole file, it
cannot know if you provided grouped or non-grouped input.

Fhirbase supports two modes (or methods) to put resources into the
database: "insert" and "copy". Insert mode uses INSERT statements and
copy mode uses COPY FROM STDIN. By default, Fhirbase uses insert mode
for local files and copy mode for Bulk Data API loads.

It does not matter for insert mode if your input is grouped or not. It
will perform with same speed on both. Use it when you're not sure what
type of input you have. Also insert mode is useful when you have
duplicate IDs in your source files (rare case but happened couple of
times). Insert mode will ignore duplicates and will persist only the
first occurrence of a specific resource instance, ignoring other
occurrences.

Copy mode is intended to be used only with grouped inputs. When
applied to grouped inputs, it's almost 3 times faster than insert
mode. But it's same slower if it's being applied to non-grouped
input.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if len(args) == 0 {
			fmt.Println("No files provided")
			return
		}

		LoadCommand(ctx, args)
		fmt.Println("done")

	},
}

type LoadConnectionConfig struct {
	Mode         string
	Numdl        uint
	Memusage     bool
	AcceptHeader string
}

func init() {
	rootCmd.AddCommand(loadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadCmd.PersistentFlags().String("foo", "", "A help for foo")
	var LoadConnectionConfig = LoadConnectionConfig{
		Mode:         "insert",
		Numdl:        5,
		Memusage:     false,
		AcceptHeader: "application/fhir+json",
	}
	loadCmd.PersistentFlags().StringVarP(&LoadConnectionConfig.Mode, "mode", "m", "insert", "insert or copy")
	loadCmd.PersistentFlags().UintVarP(&LoadConnectionConfig.Numdl, "numdl", "n", 5, "number of downloads")
	loadCmd.PersistentFlags().BoolVarP(&LoadConnectionConfig.Memusage, "memusage", "", false, "memory usage")
	loadCmd.PersistentFlags().StringVarP(&LoadConnectionConfig.AcceptHeader, "accept-header", "", "application/fhir+json", "Value for Accept HTTP header (should be application/ndjson for Cerner, application/fhir+json for Smart)")

	viper.BindPFlag("mode", loadCmd.PersistentFlags().Lookup("mode"))
	viper.BindPFlag("numdl", loadCmd.PersistentFlags().Lookup("numdl"))
	viper.BindPFlag("memusage", loadCmd.PersistentFlags().Lookup("memusage"))
	viper.BindPFlag("accept-header", loadCmd.PersistentFlags().Lookup("accept-header"))
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func openFile(fileName string) (*bundleFile, error) {
	result := new(bundleFile)

	f, err := os.OpenFile(fileName, os.O_RDONLY, 0644)

	if err != nil {
		return nil, fmt.Errorf("Error opening file: %v", err)
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
			return nil, fmt.Errorf("Error transforming resource: %v", err)
		}

		id, ok := res["id"].(string)

		if !ok {
			id = uuid.New().String()
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

	content, err := io.ReadAll(b.file)

	if err != nil {
		return nil, fmt.Errorf("Error reading file: %v", err)
	}

	iter := jsoniter.ConfigFastest.BorrowIterator(content)
	defer jsoniter.ConfigFastest.ReturnIterator(iter)

	res := iter.Read()

	if res == nil {
		return nil, fmt.Errorf("Error parsing JSON: %v", iter.Error)
	}

	resMap, ok := res.(map[string]interface{})

	if !ok {
		return nil, fmt.Errorf("got non-object value in the entries array")
	}

	b.alreadyRead = true

	return resMap, nil
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
		return nil, fmt.Errorf("cannot find `entry` key in the bundle: %v", err)
	}

	linesCount, err := countEntriesInBundle(result.iter)

	result.file.Rewind()
	result.iter.Reset(result.file)

	if err != nil {
		return nil, fmt.Errorf("cannot count entries in the bundle: %v", err)
	}

	err = goToEntriesInFhirBundle(result.iter)

	if err != nil {
		return nil, fmt.Errorf("cannot find `entry` key in the bundle: %v", err)
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
		return nil, fmt.Errorf("cannot count lines in the file: %v", err)
	}

	result.file.Rewind()

	result.count = linesCount

	return &result, nil
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

	if currentBndl == nil {
		b.currentBndlIdx = b.currentBndlIdx + 1

		return b.Next()
	}

	res, err := currentBndl.Next()

	if err != nil {
		if err == io.EOF {
			currentBndl.Close()
			b.bundles[b.currentBndlIdx] = nil
			b.currentBndlIdx = b.currentBndlIdx + 1

			return b.Next()
		}

		return nil, fmt.Errorf("Error reading resource: %v", err)
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

func (l *copyLoader) Load(ctx context.Context, db *pgxpool.Pool, bndl bundle, cb loaderCb) error {
	src := newCopyFromBundleSource(bndl, l.fhirVersion, cb)

	for src.ResourceType() != "" {
		tableName := strings.ToLower(src.ResourceType())

		_, err := db.CopyFrom(context.Background(), pgx.Identifier{tableName}, []string{"id", "txid", "status", "resource"}, src)

		if err != nil {
			return fmt.Errorf("Error copying data to %s: %v", tableName, err)
		}
	}

	return nil
}

// type JSONValue map[string]interface{}

// func (j JSONValue) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
//     jsonData, err := jsoniter.Marshal(j)
//     if err != nil {
//         return nil, err
//     }
//     return append(buf, jsonData...), nil
// }

//	func (j *JSONValue) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
//	    return jsoniter.Unmarshal(src, j)
//	}
func (l *insertLoader) Load(ctx context.Context, db *pgxpool.Pool, bndl bundle, cb loaderCb) error {
	file, err := os.OpenFile("output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	conn, err := db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("Error acquiring connection: %v", err)
	}
	defer conn.Release()
	batch := &pgx.Batch{}
	curResource := uint(0)
	totalCount := uint(bndl.Count())
	batchSize := uint(2000)

	for {
		startTime := time.Now()
		var resource map[string]interface{}
		resource, err = bndl.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("Error retrieving next resource: %v", err)
		}

		transformedResource, err := doTransform(resource, l.fhirVersion)
		if err != nil {
			fmt.Printf("Error during FB transform: %v\n", err)
			continue
		}

		resourceJSON, err := jsoniter.Marshal(transformedResource)

		if err != nil {
			return fmt.Errorf("Error marshaling transformed resource: %v", err)
		}
		resourceType, _ := resource["resourceType"].(string)
		tblName := strings.ToLower(resourceType)
		id, ok := resource["id"].(string)

		var query string
		var args []interface{} = make([]interface{}, 2)
		var numberOfArgs int = 0
		if !ok || id == "" {
			query = fmt.Sprintf(
				"INSERT INTO %s (id, txid, status, resource) VALUES (gen_random_uuid()::text, 0, 'created', $1) ON CONFLICT (id) DO NOTHING",
				tblName,
			)
			args[0] = string(resourceJSON)
			numberOfArgs = len(args)

		} else {
			query = fmt.Sprintf(
				"INSERT INTO %s (id, txid, status, resource) VALUES ($1, 0, 'created', $2) ON CONFLICT (id) DO NOTHING",
				tblName,
			)
			args[0] = id
			args[1] = string(resourceJSON)
			numberOfArgs = len(args)
			// Use provided `id` and ensure 2 placeholders are used
			// batch.Queue(
			// 	fmt.Sprintf(
			// 		"INSERT INTO %s (id, txid, status, resource) VALUES ($1, 0, 'created', $2) ON CONFLICT (id) DO NOTHING",
			// 		tblName),
			// 	[]interface{}{id, transformedResource},
			// )
		}
		// Log query and args
		// fmt.Printf("Queuing Query: %s\n", query)
		// fmt.Printf("With Arguments: %v\n", args)
		if numberOfArgs == 0 {
			return fmt.Errorf("No arguments provided for query: %s", query)
		}
		// Add to batch
		batch.Queue(query, args)
		// check to see if the this entry in the batch has an equal number of arguments to numberOfArgs

		thisQuery := batch.QueuedQueries[len(batch.QueuedQueries)-1]

		if thisQuery == nil {
			return fmt.Errorf("Error getting last query: %v", err)
		}

		if len(thisQuery.Arguments) != numberOfArgs {
			thisQuery.Arguments = args
		}

		if curResource%batchSize == 0 || curResource == totalCount-1 {
			conn, err := db.Acquire(ctx)
			if err != nil {
				return fmt.Errorf("Error acquiring connection: %v", err)
			}

			br := conn.Conn().SendBatch(ctx, batch)
			if err := br.Close(); err != nil {
				return fmt.Errorf("Error closing batch: %v", err)
			}
			batch = &pgx.Batch{}
			conn.Release()
		}

		curResource++
		cb(resourceType, time.Since(startTime))
	}

	if batch != nil {
		br := conn.SendBatch(ctx, batch)
		err := br.Close()
		if err != nil {
			return fmt.Errorf("Error closing batch: %v", err)
		}
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

func loadFiles(ctx context.Context, files []string, ldr loader, memUsage bool) error {
	database, err := db.GetConnection()
	if err != nil {
		fmt.Println("Failed to get connection config")
		return fmt.Errorf("Failed to get connection config: %v", err)
	}

	defer database.Close()
	startTime := time.Now()
	bndl, err := newMultifileBundle(files)

	if err != nil {
		return err
	}

	totalCount := bndl.Count()

	insertedCounts := make(map[string]uint)
	currentIdx := 0

	bar := progressbar.NewOptions(totalCount,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription("Loading resources..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	err = ldr.Load(ctx, database, bndl, func(curType string, duration time.Duration) {
		if memUsage && currentIdx%3000 == 0 {
			PrintMemUsage()
		}

		currentIdx = currentIdx + 1
		insertedCounts[curType] = insertedCounts[curType] + 1
		bar.Add(1)
	})

	if err != nil && err != io.EOF {

		return err
	}

	bar.Finish()
	loadDuration := int(time.Since(startTime).Seconds())

	// submitLoadEvent(insertedCounts, loadDuration)

	fmt.Printf("Done, inserted %d resources in %d seconds:\n", totalCount, loadDuration)
	fmt.Println("")

	tblw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)

	for rt, cnt := range insertedCounts {
		fmt.Fprintf(tblw, "%s\t %d\n", rt, cnt)
	}

	tblw.Flush()

	return nil
}

// LoadCommand loads FHIR schema into database
func LoadCommand(ctx context.Context, args []string) error {
	// if c.NArg() == 0 {
	// 	cli.ShowCommandHelpAndExit(ctx, c, "load", 1)
	// 	return nil
	// }

	var bulkLoad bool

	// check through the strings in args to see if any start with http
	for _, arg := range args {
		if strings.HasPrefix(arg, "http") {
			bulkLoad = true
			break
		}
	}

	fhirVersion := viper.GetString("fhir")
	mode := viper.GetString("mode")
	var ldr loader

	if bulkLoad && !viper.IsSet("mode") {
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

	memUsage := viper.GetBool("memusage")

	// if bulkLoad {
	// 	numWorkers := viper.GetInt("numdl")
	// 	acceptHdr := viper.GetString("accept-header")
	// 	fileHndlrs, err := getBulkData(args, numWorkers, acceptHdr)

	// 	if err != nil {
	// 		return err
	// 	}

	// 	files := make([]string, 0, len(fileHndlrs))

	// 	defer func() {
	// 		for _, fn := range files {
	// 			os.Remove(fn)
	// 		}
	// 	}()

	// 	for _, f := range fileHndlrs {
	// 		files = append(files, f.Name())
	// 		f.Close()
	// 	}

	// 	return loadFiles(files, ldr, memUsage)
	// }

	files, err := prewalkDirs(args)

	if err != nil {
		return fmt.Errorf("Error walking directories: %v", err)
	}

	return loadFiles(ctx, files, ldr, memUsage)
}
