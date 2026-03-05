package report

import (
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "report",
	Usage: "Generate a text report from a gedcom file",
	Subcommands: []*cli.Command{
		descendantCommand,
	},
}
