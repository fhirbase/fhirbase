package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/urfave/cli"
)

// FhirbaseRepo holds fhirbase GitHub repo name
const FhirbaseRepo = "fhirbase/fhirbase"

func updateCommand(c *cli.Context) error {
	latest, found, err := selfupdate.DetectLatest(FhirbaseRepo)

	if err != nil {
		return errors.Wrap(err, "error finding most recent fhirbase release")
	}

	currentVersion := semver.MustParse(Version)

	if !found || latest.Version.Equals(currentVersion) {
		log.Println("Current version is the latest.")
		return nil
	}

	fmt.Print("Do you want to update Fhirbase to the version", latest.Version, "? (y/n): ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')

	if err != nil || (input != "y\n" && input != "n\n") {
		fmt.Printf("Invalid input. Only 'y' or 'n' are accepted.\n")
		return nil
	}

	if input == "n\n" {
		return nil
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, os.Args[0]); err != nil {
		return errors.Wrap(err, "Error occurred while updating Fhirbase binary")
	}

	fmt.Printf("Successfully updated Fhirbase to version %s\n", latest.Version)

	return nil
}
