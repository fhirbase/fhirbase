package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

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
			Usage:       "fhirbase init <fhir-version>",
			UsageText:   "This command creates tables and other stuff to store your FHIR stuff.",
			Description: "This command creates tables and other stuff to store your FHIR stuff.",
			Action:      InitCommand,
		},
		{
			Name:        "transform",
			HelpName:    "transform",
			Hidden:      false,
			Usage:       "fhirbase transform <fhir-version> <JSON file>",
			Description: "This command transforms FHIR resource from specific file to internal FHIRBase representation and outputs result to STDOUT.",
			Action:      TransformCommand,
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
