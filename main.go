package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
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
	app.Name = "boom"
	app.Usage = "make an explosive entrance"
	app.Action = func(c *cli.Context) error {
		fmt.Println("boom! I say!")
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
