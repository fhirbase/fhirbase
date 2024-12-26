/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"embed"
	"fmt"
	"path"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	jsoniter "github.com/json-iterator/go"
	"github.com/k0kubun/go-ansi"
	db "github.com/fhirbase/fhirbase/db"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

//go:embed schema/*
var schemaFS embed.FS

// initCmd represents the init command
var initCmd = &cobra.Command{

	Use:     "init ",
	Example: "fhirbase [--fhir=FHIR version] [postgres connection options] init",
	Short:   "Creates Fhirbase schema in your database",
	Long: `
Creates SQL schema (tables, types and stored procedures) to store
resources from FHIR version specified with "--fhir" flag. Database
where schema will be created is specified with "--db" flag. Specified
database should be empty, otherwise command may fail with an SQL
error.`,

	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		InitCommand(ctx)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// PerformInit actually performs init operation
func PerformInit(db *pgxpool.Pool, fhirVersion string, cb initProgressCb) error {

	var schemaStatements []string
	var functionStatements []string
	print("in perform init")
	filename := fmt.Sprintf("fhirbase-%s.sql.json", fhirVersion)
	filepath := path.Join("schema", filename)
	schema, err := schemaFS.ReadFile(filepath)

	if err != nil {
		wrapsError := fmt.Errorf("Cannot find FHIR schema for version %s", fhirVersion)
		return wrapsError

	}
	print(schema)
	functions, err := schemaFS.ReadFile("schema/functions.sql.json")

	if err != nil {
		wrapsError := fmt.Errorf("Cannot find fhirbase function definitions")
		return wrapsError
	}
	print("schema/functions.sql.json")
	err = jsoniter.Unmarshal(schema, &schemaStatements)

	if err != nil {
		wrapsError := fmt.Errorf("Cannot parse FHIR schema '%s'", fhirVersion)
		return wrapsError
	}

	err = jsoniter.Unmarshal(functions, &functionStatements)

	if err != nil {
		wrapsError := fmt.Errorf("Cannot parse function definitions")
		return wrapsError
	}

	allStmts := append(schemaStatements, functionStatements...)
	allStmts = append(allStmts, conceptsTables...)

	conn, err := db.Acquire(context.Background())
	if err != nil {
		wrapsError := fmt.Errorf("Cannot acquire connection to database")
		return wrapsError
	}
	defer conn.Release()
	t := time.Now()
	for i, stmt := range allStmts {
		fmt.Printf("Executing statement %d of %d: %s\n", i+1, len(allStmts), stmt)
		_, err = conn.Exec(context.Background(), stmt)

		if err != nil {
			wrapsError := fmt.Errorf("PG error while executing statement:\n%s\n", stmt)
			return wrapsError
		}

		cb(i, int64(len(allStmts)), time.Since(t))

		t = time.Now()
	}

	return nil
}

// InitCommand loads FHIR schema into database
func InitCommand(ctx context.Context) {

	fmt.Println("Init called")
	fmt.Println(ctx)
	fhirVersion := viper.GetString("fhir")
	if fhirVersion == "" {
		fmt.Println("FHIR version is not specified")
		return
	}

	dbUrl := viper.GetString("db")
	if dbUrl == "" {
		fmt.Println("Database URL is not specified")
		return
	}

	conn, err := db.GetPgxConnectionConfig()

	if err != nil {
		fmt.Println("Failed to get connection config")
		return
	}

	database, err := pgxpool.NewWithConfig(ctx, conn)
	defer database.Close()

	bar := progressbar.NewOptions(100, progressbar.OptionSetWriter(ansi.NewAnsiStdout()), progressbar.OptionShowBytes(true))

	err = PerformInit(database, fhirVersion, func(curIdx int, total int64, duration time.Duration) {
		if curIdx%10 == 0 {
			bar.Add(10)
		}

		if int64(curIdx) == total-(int64(1)) {
			bar.Finish()
		}

		fmt.Printf("Statement %d of %d executed in %s\n", curIdx+1, total, duration)
	})

	if err != nil {
		fmt.Println("Failed to perform init")
		return
	}

	fmt.Printf("Database initialized with FHIR schema version '%s'\n", fhirVersion)

	return

}
