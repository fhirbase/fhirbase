package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func UnknownTotalCounter(unit int, pairFormat string, wcc ...decor.WC) decor.Decorator {
	var wc decor.WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.Init()
	d := &unknownTotalDecorator{
		WC:         wc,
		unit:       unit,
		pairFormat: pairFormat,
	}
	return d
}

type unknownTotalDecorator struct {
	WC          decor.WC
	unit        int
	pairFormat  string
	completeMsg *string
}

func (d *unknownTotalDecorator) Decor(st *decor.Statistics) string {
	if st.Completed && d.completeMsg != nil {
		return d.WC.FormatMsg(*d.completeMsg)
	}

	var str string
	switch d.unit {
	case decor.UnitKiB:
		str = fmt.Sprintf(d.pairFormat, decor.CounterKiB(st.Current))
	case decor.UnitKB:
		str = fmt.Sprintf(d.pairFormat, decor.CounterKB(st.Current))
	default:
		str = fmt.Sprintf(d.pairFormat, st.Current)
	}

	return d.WC.FormatMsg(str)
}

func (d *unknownTotalDecorator) Syncable() (bool, chan int) {
	return d.WC.Syncable()
}

var matchNonDigits, _ = regexp.Compile("[^\\d]")

func getBulkDataFiles(pingURL string) ([]string, error) {
	fmt.Println("Waiting for Bulk Data API server to prepare files...")

	resp, err := http.Get(pingURL)

	if err != nil {
		return nil, errors.Wrap(err, "error while pinging Bulk Data API server")
	}

	for i := 1; resp.StatusCode != 200; i++ {
		// progress := int64(-1)
		// xprogress := matchNonDigits.ReplaceAllString(resp.Header.Get("X-Progress"), "")

		// if len(xprogress) > 0 {
		// 	progress, err = strconv.ParseInt(xprogress, 10, 64)

		// 	if err != nil {
		// 		progress = -1
		// 	}
		// }

		if resp.StatusCode > 299 || resp.StatusCode < 200 {
			respBody, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				return nil, fmt.Errorf("got %d response wile piging Bulk Data API server; cannot read response body", resp.StatusCode)
			}

			return nil, fmt.Errorf("got %d response wile piging Bulk Data API server. Response Body:\n%s", resp.StatusCode, respBody)
		}

		if i%5 == 0 {
			fmt.Println("still waiting...")
		}

		resp.Body.Close()

		time.Sleep(1000 * time.Millisecond)

		resp, err = http.Get(pingURL)

		if err != nil {
			return nil, errors.Wrap(err, "error while pinging Bulk Data API server")
		}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.Wrap(err, "error while reading response")
	}

	iter := jsoniter.ConfigDefault.BorrowIterator(body)
	defer jsoniter.ConfigDefault.ReturnIterator(iter)

	obj := iter.Read()

	if obj == nil {
		return nil, errors.Wrap(iter.Error, "cannot parse server response")
	}

	fileURLs := make([]string, 0)
	objMap, ok := obj.(map[string]interface{})

	if !ok {
		return nil, fmt.Errorf("expecting JSON object at the top level")
	}

	output := objMap["output"]

	if output == nil {
		return nil, fmt.Errorf("expecting to have 'output' attribute")
	}

	outputArr, ok := output.([]interface{})

	if !ok {
		return nil, fmt.Errorf("'output' attribute is not an JSON Array")
	}

	for _, v := range outputArr {
		item, ok := v.(map[string]interface{})

		if !ok {
			return nil, fmt.Errorf("got non-object in 'output' array")
		}

		url := item["url"]

		if url == nil {
			return nil, fmt.Errorf("cannot get 'url' attribute in item of 'output' array")
		}

		urlString, ok := url.(string)

		if !ok {
			return nil, fmt.Errorf("'url' attribute is not a string")
		}

		fileURLs = append(fileURLs, urlString)
	}

	return fileURLs, nil
}

func stripURL(url string, length int) string {
	if len(url) < length {
		return strings.Repeat(" ", length-len(url)) + url
	}

	return "..." + url[len(url)-length-3:len(url)]
}

func startDlWorker(n int, bars *mpb.Progress, jobs chan string, results chan interface{}, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		for url := range jobs {
			targetFile, err := ioutil.TempFile("", "")

			if err != nil {
				results <- errors.Wrap(err, "cannot create temp file")
				continue
			}

			resp, err := http.Get(url)

			if err != nil {
				results <- errors.Wrap(err, "cannot perform HTTP request")
				continue
			}

			if resp.StatusCode != 200 {
				results <- fmt.Errorf("got non-200 response while downloading %s", url)
			}

			contentLengthHeader := resp.Header.Get("Content-Length")
			size, err := strconv.ParseInt(contentLengthHeader, 10, 64)

			counterDecorator := decor.CountersKibiByte("%6.1f / %6.1f", decor.WCSyncWidth)

			if err != nil {
				size = 0
				counterDecorator = UnknownTotalCounter(decor.UnitKiB, "%6.1f / ???", decor.WCSyncWidth)
			}

			name := stripURL(url, 25)
			bar := bars.AddBar(size, mpb.BarPriority(n),
				mpb.BarRemoveOnComplete(),
				mpb.PrependDecorators(
					decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
					counterDecorator,
				),
				mpb.AppendDecorators(
					decor.EwmaETA(decor.ET_STYLE_MMSS, 1024*4, decor.WCSyncWidth),
					decor.AverageSpeed(decor.UnitKiB, " % .2f"),
				),
			)

			reader := bar.ProxyReader(resp.Body)

			totalWritten, err := io.Copy(targetFile, reader)

			if err != nil {
				results <- errors.Wrap(err, "cannot copy HTTP response body to temporary file")
			}

			bar.SetTotal(totalWritten, true)

			results <- targetFile
		}
	}()
}

func downloadFiles(fileURLs []string, numWorkers int) ([]*os.File, error) {
	doneWg := new(sync.WaitGroup)
	bars := mpb.New(mpb.WithWidth(64), mpb.WithWaitGroup(doneWg))
	jobs := make(chan string, len(fileURLs))
	results := make(chan interface{}, len(fileURLs))
	files := make([]*os.File, 0)

	fmt.Printf("Start downloading %d files in %d threads\n", len(fileURLs), numWorkers)

	for _, url := range fileURLs {
		jobs <- url
	}

	close(jobs)

	for i := 0; i < numWorkers; i++ {
		startDlWorker(i, bars, jobs, results, doneWg)
	}

	bars.Wait()

	close(results)

	for res := range results {
		err, ok := res.(error)

		if ok {
			fmt.Printf("Got an error while downloading file: %s", err.Error())
		} else {
			f, ok := res.(*os.File)

			if ok {
				files = append(files, f)
			} else {
				fmt.Printf("got result of unknown type: %v", res)
			}
		}
	}

	fmt.Printf("Finished downloading, got %d files\n", len(files))

	return files, nil
}

func getBulkData(url string) ([]*os.File, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Prefer", "respond-async")
	req.Header.Add("Accept", "application/fhir+json")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if err != nil {
		return nil, errors.Wrap(err, "cannot perform HTTP query")
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("expected 200 response, got %d", resp.StatusCode)
	}

	pingURL := resp.Header.Get("Content-Location")

	if len(pingURL) == 0 {
		return nil, fmt.Errorf("No Content-Location header was returned by Bulk Data API server")
	}

	fileURLs, err := getBulkDataFiles(pingURL)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot get list of files to download")
	}

	return downloadFiles(fileURLs, 5)
}
