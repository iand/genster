/*
This is free and unencumbered software released into the public domain. For more
information, see <http://unlicense.org/> or the accompanying UNLICENSE file.
*/

package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/iand/genster/chart"
	"github.com/iand/genster/site"
)

func main() {
	app := &cli.App{
		Name:     "genster",
		HelpName: "genster",
		Usage:    "Generate a website from a gedcom file",
		Commands: []*cli.Command{
			site.Command,
			chart.Command,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
