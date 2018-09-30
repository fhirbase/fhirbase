package main

import (
	"fmt"
	"log"

	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/urfave/cli"
)

// FhirbaseRepo holds fhirbase GitHub repo name
const FhirbaseRepo = "fhirbase/fhirbase"

func updateCommand(c *cli.Context) error {
	latest, found, err := selfupdate.DetectLatest(FhirbaseRepo)
	if err != nil {
		log.Println("Error occurred while detecting version:", err)
		return err
	}

	fmt.Printf("Latest fhirbase version %v %v:", found, latest)

	return nil
}
