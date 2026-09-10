/*
This is free and unencumbered software released into the public domain. For more
information, see <http://unlicense.org/> or the accompanying UNLICENSE file.
*/

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/iand/genster/annotate"
	"github.com/iand/genster/build"
	"github.com/iand/genster/chart"
	"github.com/iand/genster/report"
	"github.com/iand/genster/serve"
	"github.com/iand/genster/site"
)

func main() {
	app := &cli.Command{
		Name:  "genster",
		Usage: "Generate a website from a gedcom file",
		Commands: []*cli.Command{
			site.Command,
			build.Command,
			serve.Command,
			chart.Command,
			report.Command,
			annotate.Command,
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
