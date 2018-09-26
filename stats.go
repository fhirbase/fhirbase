package main

import (
	"github.com/denisbrodbeck/machineid"
)

func getMachineID() (string, error) {
	return machineid.ProtectedID("fhirbase")
}
