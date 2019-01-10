package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/blang/semver"
	update "github.com/inconshreveable/go-update"
	"github.com/pkg/errors"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/urfave/cli"
)

// FhirbaseRepo holds fhirbase GitHub repo name
const FhirbaseRepo = "fhirbase/fhirbase"

func readYesNo() bool {
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')

	input = strings.ToLower(input)

	if err != nil || (input != "y\n" && input != "n\n" && input != "yes\n" && input != "no\n") {
		fmt.Printf("Invalid input. Only 'y', 'n', 'yes', 'no'j are accepted.\n")
		return false
	}

	if input == "y\n" || input == "yes\n" {
		return true
	}

	return false
}

func updateStableBuild() error {
	selfupdate.EnableLog()

	latest, found, err := selfupdate.DetectLatest(FhirbaseRepo)

	if err != nil {
		return errors.Wrap(err, "error finding most recent fhirbase release")
	}

	currentVersion := semver.MustParse(Version[1:len(Version)])

	if !found || latest.Version.Equals(currentVersion) {
		fmt.Printf("Current version is the latest.\n")
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

	url := fmt.Sprintf("https://github.com/fhirbase/fhirbase/releases/download/nightly-build/fhirbase-%s-%s", runtime.GOOS, runtime.GOARCH)

	resp, err := http.Get(url)

	if err != nil {
		return errors.Wrap(err, "cannot perform http query to get latest nightly build")
	}

	defer resp.Body.Close()
	fmt.Printf("HTTP request finished with %d, starting update...\n", resp.StatusCode)

	err = update.Apply(resp.Body, update.Options{})

	if err != nil {
		return errors.Wrap(err, "cannot apply Fhirbase update")
	}

	fmt.Printf("Fhirbase updated!\n")

	return nil
}

func updateCommand(c *cli.Context) error {
	if strings.HasPrefix(Version, "nightly") {
		return updateNightlyBuild()
	}

	return updateStableBuild()
}
