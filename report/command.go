package report

import (
	"github.com/urfave/cli/v3"
)

var Command = &cli.Command{
	Name:  "report",
	Usage: "Generate a text report from a gedcom file",
	Commands: []*cli.Command{
		descendantCommand,
	},
}
