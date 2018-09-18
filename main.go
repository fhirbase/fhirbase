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
	app.Usage = "command-line utility to operate on FHIR data with PostgreSQL database."

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
			Usage:       "Creates FHIRbase schema in specific database",
			UsageText:   "fhirbase [--fhir=FHIR version] [postgres connection options] init",
			Description: "Creates SQL to store FHIR resources for specified FHIR version, as well as stored procedures for CRUD operations. Database should be empty, otherwise this command will fail with an error.",
			Action:      InitCommand,
		},
		{
			Name:        "transform",
			HelpName:    "transform",
			Hidden:      false,
			Usage:       "Performs FHIRbase transformation on a single FHIR resource loaded from JSON file",
			UsageText:   "fhirbase [--fhir=FHIR version] transform path/to/fhir-resource.json",
			Description: "This command transforms FHIR resource from specific file into internal FHIRbase representation and outputs result to STDOUT.",
			Action:      TransformCommand,
		},
		{
			Name:        "bulkget",
			HelpName:    "bulkget",
			Hidden:      false,
			ArgsUsage:   "[BULK DATA ENDPOINT] [TARGET DIR]",
			Usage:       "Downloads FHIR data from Bulk Data API endpoint and save it on local filesystem",
			UsageText:   "fhirbase bulkget --numdl=10 http://some-fhir-server.com/fhir/Patient/$everything ./output-dir/",
			Description: "Downloads FHIR data from Bulk Data API endpoint and saves results into specific directory on local filesystem.",
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
			Usage:       "Imports FHIR resources into specific database",
			ArgsUsage:   "[BULK DATA URL OR FILE PATHS]",
			Description: "This command loads FHIR data from various sources, i.e. local file or Bulk Data API server.",
			Action:      LoadCommand,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "mode, m",
					Value: "copy",
					Usage: "Method how data import will be performed. Possible values: 'copy' or 'insert'",
				},
				cli.UintFlag{
					Name:  "numdl",
					Value: 5,
					Usage: "Number of parallel downloads for Bulk Data API client",
				},
				cli.BoolFlag{
					Name:  "memusage",
					Usage: "Outputs memory usage during resources loading (for debug purposes)",
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
			Usage:       "Starts web server with primitive UI to perform SQL queries from the browser",
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
					Value: "",
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
