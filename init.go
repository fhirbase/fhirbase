package main

import (
	"fmt"

	"github.com/gobuffalo/packr"
	"github.com/jackc/pgx"
	jsoniter "github.com/json-iterator/go"
	"github.com/urfave/cli"

	"log"
	"os"
)

// AvailableSchemas contains all know FHIR versions
var AvailableSchemas = []string{
	"1.0.2", "1.1.0", "1.4.0",
	"1.6.0", "1.8.0", "3.0.1",
	"3.2.0", "3.3.0", "dev",
}

// PerformInit actually performs init operation
func PerformInit(db *pgx.Conn, fhirVersion string) error {
	var schemaStatements []string
	var functionStatements []string

	box := packr.NewBox("./data")
	schema, err := box.MustBytes(fmt.Sprintf("schema/fhirbase-%s.sql.json", fhirVersion))

	if err != nil {
		log.Fatalf("Cannot find FHIR schema '%s'", fhirVersion)
	}

	functions, err := box.MustBytes("schema/functions.sql.json")

	if err != nil {
		log.Fatalf("Cannot find fhirbase function definitions: %v", err)
	}

	err = jsoniter.Unmarshal(schema, &schemaStatements)

	if err != nil {
		log.Fatalf("Cannot parse FHIR schema '%s': %v", fhirVersion, err)
	}

	err = jsoniter.Unmarshal(functions, &functionStatements)

	if err != nil {
		log.Fatalf("Cannot parse function definitions: %v", err)
	}

	for _, stmt := range append(schemaStatements, functionStatements...) {
		_, err = db.Exec(stmt)

		if err != nil {
			log.Printf("PG error: %v\nWhile executing statement:\n%s\n", err, stmt)
		}
	}

	return nil
}

// InitCommand loads FHIR schema into database
func InitCommand(c *cli.Context) error {
	var fhirVersion string

	if c.NArg() > 0 {
		fhirVersion = c.Args().Get(0)
	} else {
		log.Printf("You must provide a FHIR version for `fhirbase init` command.\nKnow FHIR versions are: %v", AvailableSchemas)
		os.Exit(1)
	}

	db := GetConnection(nil)
	PerformInit(db, fhirVersion)

	log.Printf("Database initialized with FHIR schema version '%s'", fhirVersion)

	return nil
}
