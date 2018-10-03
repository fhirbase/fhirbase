package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/denisbrodbeck/machineid"
	jsoniter "github.com/json-iterator/go"
)

const EventEndpointURL = "https://license.aidbox.app/event"

var eventsWg sync.WaitGroup

// DisableStats flag disables stats submition
var DisableStats = false

func getMachineID() (string, error) {
	return machineid.ProtectedID("fhirbase")
}

func submitEvent(et string, payload interface{}) {
	if DisableStats {
		return
	}

	eventsWg.Add(1)
	defer eventsWg.Done()

	// fmt.Printf("event: submitting %s event\n", et)

	machineID, err := getMachineID()

	// fmt.Printf("event: machine id %s %s\n", machineID, err)

	if err != nil {
		fmt.Printf("stats error: %+v\n", err)
		return
	}

	body, err := jsoniter.MarshalToString(map[string]interface{}{
		"product":   "fhirbase",
		"machineId": machineID,
		"version":   Version,
		"type":      et,
		"payload": map[string]interface{}{
			"os":    runtime.GOOS,
			"arch":  runtime.GOARCH,
			"event": payload,
		},
	})

	// fmt.Printf("event: marshalled stuff %s\n", body)

	if err != nil {
		// fmt.Printf("stats error: %+v\n", err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", EventEndpointURL, bytes.NewBuffer([]byte(body)))

	if err != nil {
		// fmt.Printf("stats error: %+v\n", err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		// fmt.Printf("stats error: %+v\n", err)
		return
	}

	// fmt.Printf("event: submitted\n")

	defer resp.Body.Close()
}

func submitInitEvent(fhirVersion string) {
	go submitEvent("init", map[string]string{
		"fhirVersion": fhirVersion,
	})
}

func submitErrorEvent(err error) {
	go submitEvent("error", map[string]string{
		"error": fmt.Sprintf("%+v", err),
	})
}

func submitLoadEvent(stats map[string]uint, drtion int) {
	go submitEvent("load", map[string]interface{}{
		"stats":    stats,
		"duration": drtion,
	})
}

func waitForAllEventsSubmitted() {
	eventsWg.Wait()
}
