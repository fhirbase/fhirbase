package main

import (
	"testing"
)

func TestMachineId(t *testing.T) {
	firstId, err := getMachineID()

	if err != nil {
		t.Error(err)
	}

	secondId, err := getMachineID()

	if err != nil {
		t.Error(err)
	}

	if firstId != secondId {
		t.Log("Expected to get two identical machine IDs, got different")
		t.Fail()
	}
}
