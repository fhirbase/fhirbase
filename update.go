package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/urfave/cli"
)

// FhirbaseRepo holds fhirbase GitHub repo name
const FhirbaseRepo = "fhirbase/fhirbase"

func readYesNo() bool {
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')

	if err != nil || (input != "y\n" && input != "n\n" && input != "Y\n" && input != "N\n") {
		fmt.Printf("Invalid input. Only 'y' or 'n' are accepted.\n")
		return false
	}

	if input == "y\n" || input == "Y\n" {
		return true
	}

	return false

}

func updateStableBuild() error {
	latest, found, err := selfupdate.DetectLatest(FhirbaseRepo)

	if err != nil {
		return errors.Wrap(err, "error finding most recent fhirbase release")
	}

	currentVersion := semver.MustParse(Version[1:len(Version)])

	if !found || latest.Version.Equals(currentVersion) {
		log.Println("Current version is the latest.")
		return nil
	}

	fmt.Printf("Do you want to update Fhirbase to the version %s? (yes/no): ", latest.Version)

	if !readYesNo() {
		return nil
	}

	fmt.Printf("Updating...\n")

	if err := selfupdate.UpdateTo(latest.AssetURL, os.Args[0]); err != nil {
		return errors.Wrap(err, "Error occurred while updating Fhirbase binary")
	}

	fmt.Printf("Successfully updated Fhirbase to the version %s.\n", latest.Version)

	return nil
}

func updateNightlyBuild() error {
	fmt.Print("Do you want to update Fhirbase to the latest nightly (unstable) build? (y/n): ")
	if !readYesNo() {
		return nil
	}

	fmt.Printf("TODO: not implemented yet\n")

	return nil
}

func updateCommand(c *cli.Context) error {
	if strings.HasPrefix(Version, "nightly") {
		return updateNightlyBuild()
	}

	return updateStableBuild()
}
