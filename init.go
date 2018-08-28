package main

import (
	"fmt"

	"github.com/gobuffalo/packr"
	jsoniter "github.com/json-iterator/go"
	"github.com/urfave/cli"

	"log"
	"os"
)

// InitCommand loads FHIR schema into database
func InitCommand(c *cli.Context) error {
	var fhirVersion string
	var schemaStatements []string

	if c.NArg() > 0 {
		fhirVersion = c.Args().Get(0)
	} else {
		log.Printf("You must provide a FHIR version for `fhirbase init` command.")
		os.Exit(1)
	}

	box := packr.NewBox("./data")
	schema, err := box.MustBytes(fmt.Sprintf("schema/fhirbase-%s.sql.json", fhirVersion))

	if err != nil {
		log.Fatalf("Cannot find FHIR schema '%s'", fhirVersion)
	}

	err = jsoniter.Unmarshal(schema, &schemaStatements)

	if err != nil {
		log.Fatalf("Cannot parse FHIR schema '%s': %v", fhirVersion, err)
	}

	db := GetConnection()

	for _, stmt := range schemaStatements {
		_, err = db.Exec(stmt)

		if err != nil {
			log.Printf("PG error: %v\nWhile executing statement:\n%s\n", err, stmt)
		}
	}

	log.Printf("Database initialized with FHIR schema version '%s'", fhirVersion)

	return nil
}
