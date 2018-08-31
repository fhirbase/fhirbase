package fhirbase

import (
	"fmt"

	"github.com/gobuffalo/packr"
	"github.com/jackc/pgx"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"log"
)

// PerformInit actually performs init operation
func PerformInit(db *pgx.Conn, fhirVersion string) error {
	var schemaStatements []string
	var functionStatements []string

	box := packr.NewBox("./schema")
	schema, err := box.MustBytes(fmt.Sprintf("fhirbase-%s.sql.json", fhirVersion))

	if err != nil {
		return errors.Wrapf(err, "Cannot find FHIR schema '%s'", fhirVersion)
	}

	functions, err := box.MustBytes("functions.sql.json")

	if err != nil {
		return errors.Wrap(err, "Cannot find fhirbase function definitions")
	}

	err = jsoniter.Unmarshal(schema, &schemaStatements)

	if err != nil {
		return errors.Wrapf(err, "Cannot parse FHIR schema '%s'", fhirVersion)
	}

	err = jsoniter.Unmarshal(functions, &functionStatements)

	if err != nil {
		return errors.Wrap(err, "Cannot parse function definitions")
	}

	for _, stmt := range append(schemaStatements, functionStatements...) {
		_, err = db.Exec(stmt)

		if err != nil {
			return errors.Wrapf(err, "PG error while executing statement:\n%s\n", stmt)
		}
	}

	return nil
}

// InitCommand loads FHIR schema into database
func InitCommand(c *cli.Context) error {
	fhirVersion := c.GlobalString("fhir")

	db := GetConnection(nil)
	err := PerformInit(db, fhirVersion)

	if err != nil {
		return errors.Wrap(err, "Failed to perform init command. Perhaps target database is not empty?")
	}

	log.Printf("Database initialized with FHIR schema version '%s'", fhirVersion)

	return nil
}
