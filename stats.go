package main

import (
	"bytes"
	"net/http"

	"github.com/denisbrodbeck/machineid"
	jsoniter "github.com/json-iterator/go"
)

const EventEndpointURL = "https://license.aidbox.app/event"

func getMachineID() (string, error) {
	return machineid.ProtectedID("fhirbase")
}

func submitEvent(et string, payload interface{}) {
	machineID, err := getMachineID()

	if err == nil {
		return
	}

	body, err := jsoniter.MarshalToString(map[string]interface{}{
		"product":   "fhirbase",
		"machineId": machineID,
		"version":   "0.0.1",
		"type":      et,
		"payload":   payload,
	})

	client := &http.Client{}
	req, err := http.NewRequest("POST", EventEndpointURL, bytes.NewBuffer([]byte(body)))

	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()
}

func submitInitEvent(fhirVersion string) {
	submitEvent("init", map[string]string{
		"fhirVersion": fhirVersion,
	})
}
