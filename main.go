package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli"
)

// AvailableSchemas contains all know FHIR versions
var AvailableSchemas = []string{
	"1.0.2", "1.1.0", "1.4.0",
	"1.6.0", "1.8.0", "3.0.1",
	"3.2.0", "3.3.0", "dev",
}

const logo = ` (        )  (    (                   (
 )\ )  ( /(  )\ ) )\ )   (     (      )\ )
(()/(  )\())(()/((()/( ( )\    )\    (()/( (
 /(_))((_)\  /(_))/(_)))((_)((((_)(   /(_)))\
(_))_| _((_)(_)) (_)) ((_)_  )\ _ )\ (_)) ((_)
| |_  | || ||_ _|| _ \ | _ ) (_)_\(_)/ __|| __|
| __| | __ | | | |   / | _ \  / _ \  \__ \| _|
|_|   |_||_||___||_|_\ |___/ /_/ \_\ |___/|___|        v1.0`

func main() {
	cli.AppHelpTemplate = fmt.Sprintf("%s\n\n%s", logo, cli.AppHelpTemplate)

	app := cli.NewApp()
	app.Name = "fhirbase"
	app.Usage = "command-line tool to create fhirbase schema and import FHIR data"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "host, n",
			Value:       "localhost",
			Usage:       "PostgreSQL host",
			EnvVar:      "PGHOST",
			Destination: &PgConfig.Host,
		},
		cli.UintFlag{
			Name:        "port, p",
			Value:       5432,
			Usage:       "PostgreSQL port",
			EnvVar:      "PGPORT",
			Destination: &PgConfig.Port,
		},
		cli.StringFlag{
			Name:        "username, U",
			Value:       "postgres",
			Usage:       "PostgreSQL username",
			EnvVar:      "PGUSER",
			Destination: &PgConfig.Username,
		},
		cli.StringFlag{
			Name:  "fhir, f",
			Value: "3.3.0",
			Usage: "FHIR version to use. Know FHIR versions are: " + strings.Join(AvailableSchemas, ", "),
		},
		cli.StringFlag{
			Name:        "db, d",
			Value:       "",
			Usage:       "Database to connect to",
			EnvVar:      "PGDATABASE",
			Destination: &PgConfig.Database,
		},
		cli.StringFlag{
			Name:        "password, W",
			Usage:       "PostgreSQL password",
			EnvVar:      "PGPASSWORD",
			Destination: &PgConfig.Password,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "init",
			HelpName:    "init",
			Hidden:      false,
			Usage:       "Creates FHIRBase schema in Postgres database",
			UsageText:   "This command creates tables and other stuff to store your FHIR stuff.",
			Description: "This command creates tables and other stuff to store your FHIR stuff.",
			Action:      InitCommand,
		},
		{
			Name:        "transform",
			HelpName:    "transform",
			Hidden:      false,
			Usage:       "Performs FHIRBase transformation on a single resource from JSON file",
			Description: "This command transforms FHIR resource from specific file to internal FHIRBase representation and outputs result to STDOUT.",
			Action:      TransformCommand,
		},
		{
			Name:        "bulkget",
			HelpName:    "bulkget",
			Hidden:      false,
			ArgsUsage:   "[BULK DATA ENDPOINT] [TARGET DIR]",
			Usage:       "Downloads FHIR data from Bulk Data API endpoint",
			Description: "Downloads FHIR data from Bulk Data API endpoint.",
			Action:      BulkGetCommand,
			Flags: []cli.Flag{
				cli.UintFlag{
					Name:  "numdl",
					Value: 5,
					Usage: "Number of parallel downloads for Bulk Data API client",
				},
				cli.StringFlag{
					Name:  "accept-header",
					Value: "application/fhir+json",
					Usage: "Value for Accept HTTP header (should be application/ndjson for Cerner, application/fhir+json for Smart)",
				},
			},
		},
		{
			Name:        "load",
			HelpName:    "load",
			Hidden:      false,
			Usage:       "Loads FHIR data (resources) into database",
			ArgsUsage:   "[BULK DATA URL OR FILE PATHS]",
			Description: "This command loads FHIR data from various sources, i.e. local file or Bulk Data API server.",
			Action:      LoadCommand,
			Flags: []cli.Flag{
				cli.UintFlag{
					Name:  "batchsize, b",
					Value: 2000,
					Usage: "Number of INSERTs to send in one query",
				},
				cli.UintFlag{
					Name:  "numdl",
					Value: 5,
					Usage: "Number of parallel downloads for Bulk Data API client",
				},
				cli.StringFlag{
					Name:  "accept-header",
					Value: "application/fhir+json",
					Usage: "Value for Accept HTTP header (should be application/ndjson for Cerner, application/fhir+json for Smart)",
				},
			},
		},
		{
			Name:        "web",
			HelpName:    "web",
			Hidden:      false,
			Usage:       "Starts simple web server to invoke SQL queries from browser",
			ArgsUsage:   "",
			Description: "Starts simple web server to invoke SQL queries from browser",
			Action:      WebCommand,
			Flags: []cli.Flag{
				cli.UintFlag{
					Name:  "webport",
					Value: 3000,
					Usage: "Port to start webserver on",
				},
				cli.StringFlag{
					Name:  "webhost",
					Value: "localhost",
					Usage: "Host to start webserver on",
				},
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		cli.HelpPrinter(os.Stdout, cli.AppHelpTemplate, app)
		return nil
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
