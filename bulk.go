package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

var matchNonDigits, _ = regexp.Compile("[^\\d]")

func waitForBulkData(pingURL string) ([]string, error) {
	resp, err := http.Get(pingURL)

	for resp.StatusCode != 200 {
		progress := int64(-1)
		xprogress := matchNonDigits.ReplaceAllString(resp.Header.Get("X-Progress"), "")

		if len(xprogress) > 0 {
			progress, err = strconv.ParseInt(xprogress, 10, 64)

			if err != nil {
				progress = -1
			}
		}

		fmt.Printf("bulk prepare progress: %d\n", progress)
		resp.Body.Close()
		time.Sleep(1000 * time.Millisecond)
		resp, err = http.Get(pingURL)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	iter := jsoniter.ConfigDefault.BorrowIterator(body)
	defer jsoniter.ConfigDefault.ReturnIterator(iter)

	obj := iter.Read()
	fileURLs := make([]string, 0)

	if obj != nil {
		objMap, ok := obj.(map[string]interface{})

		if ok {
			for _, v := range objMap["output"].([]interface{}) {
				item, ok := v.(map[string]interface{})

				if ok {
					fileURLs = append(fileURLs, item["url"].(string))
				}
			}
		}
	}

	return fileURLs, nil
}

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

func startDlWorker(n int, bars *mpb.Progress, jobs chan string, results chan string, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		for url := range jobs {
			targetFile, _ := ioutil.TempFile("", "")
			resp, _ := http.Get(url)

			contentLengthHeader := resp.Header.Get("Content-Length")
			size, err := strconv.ParseInt(contentLengthHeader, 10, 64)
			counterDecorator := decor.CountersKibiByte("%6.1f / %6.1f", decor.WCSyncWidth)

			if err != nil {
				size = 0
				counterDecorator = UnknownTotalCounter(decor.UnitKiB, "%6.1f / ???", decor.WCSyncWidth)
			}

			name := "..." + url[len(url)-20:len(url)]

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

			totalWritten, _ := io.Copy(targetFile, reader)

			bar.SetTotal(totalWritten, true)

			results <- targetFile.Name()
		}
	}()
}

func downloadFiles(fileURLs []string, numWorkers int) error {
	doneWg := new(sync.WaitGroup)
	bars := mpb.New(mpb.WithWidth(64), mpb.WithWaitGroup(doneWg))
	jobs := make(chan string, len(fileURLs))
	results := make(chan string, len(fileURLs))

	for _, url := range fileURLs {
		jobs <- url
	}

	close(jobs)

	for i := 0; i < numWorkers; i++ {
		startDlWorker(i, bars, jobs, results, doneWg)
	}

	bars.Wait()

	fmt.Printf("%v", results)

	return nil
}

func getBulkData(url string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Prefer", "respond-async")
	req.Header.Add("Accept", "application/fhir+json")
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 220 {
		pingUrl := resp.Header.Get("Content-Location")
		fileURLs, _ := waitForBulkData(pingUrl)

		fmt.Printf("HERE!\n")

		downloadFiles(fileURLs, 5)
	}

	// fmt.Printf("%v\n%s\n", resp, body)

	return nil
}
