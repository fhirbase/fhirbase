package main

import (
	"fmt"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/jackc/pgx"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

var conceptsTables = []string{
	`CREATE TABLE IF NOT EXISTS "concept" (
id text primary key,
txid bigint not null,
ts timestamptz DEFAULT current_timestamp,
resource_type text default 'Concept',
status resource_status not null,
resource jsonb not null);`,

	`CREATE TABLE IF NOT EXISTS "concept_history" (
id text,
txid bigint not null,
ts timestamptz DEFAULT current_timestamp,
resource_type text default 'Concept',
status resource_status not null,
resource jsonb not null,
PRIMARY KEY (id, txid)
);`}

type initProgressCb func(curIdx int, total int64, duration time.Duration)

// PerformInit actually performs init operation
func PerformInit(db *pgx.Conn, fhirVersion string, cb initProgressCb) error {
	var schemaStatements []string
	var functionStatements []string

	box := packr.NewBox("./schema")
	schema, err := box.Find(fmt.Sprintf("fhirbase-%s.sql.json", fhirVersion))

	if err != nil {
		return errors.Wrapf(err, "Cannot find FHIR schema '%s'", fhirVersion)
	}

	functions, err := box.Find("functions.sql.json")

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

	allStmts := append(schemaStatements, functionStatements...)
	allStmts = append(allStmts, conceptsTables...)

	t := time.Now()
	for i, stmt := range allStmts {
		_, err = db.Exec(stmt)

		if err != nil {
			return errors.Wrapf(err, "PG error while executing statement:\n%s\n", stmt)
		}

		cb(i, int64(len(allStmts)), time.Since(t))
		t = time.Now()
	}

	return nil
}

// InitCommand loads FHIR schema into database
func InitCommand(c *cli.Context) error {
	fhirVersion := c.GlobalString("fhir")
	db := GetConnection(nil)

	bars := mpb.New(
		mpb.WithWidth(100),
	)

	bar := bars.AddBar(int64(1),
		mpb.AppendDecorators(
			decor.Percentage(decor.WC{W: 3}),
			decor.AverageETA(decor.ET_STYLE_MMSS, decor.WC{W: 6}),
		),
		mpb.PrependDecorators(decor.CountersNoUnit("%d / %d", decor.WC{W: 10})))

	err := PerformInit(db, fhirVersion, func(curIdx int, total int64, duration time.Duration) {
		if curIdx == 0 {
			bar.SetTotal(total, false)
		}

		bar.IncrBy(1, duration)
	})

	if err != nil {
		bars.Abort(bar, false)
		return errors.Wrap(err, "Failed to perform init command. Perhaps target database is not empty?")
	}

	bars.Wait()

	submitInitEvent(fhirVersion)

	fmt.Printf("Database initialized with FHIR schema version '%s'\n", fhirVersion)

	return nil
}
