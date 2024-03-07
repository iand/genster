package logging

import (
	"log/slog"

	"github.com/iand/pontium/hlog"
	"github.com/kortschak/utter"
	"github.com/urfave/cli/v2"
)

var Flags = []cli.Flag{
	&cli.BoolFlag{
		Name:        "verbose",
		Aliases:     []string{"v"},
		Usage:       "Set logging level more verbose to include info level logs",
		Value:       false,
		Destination: &Opts.Verbose,
	},

	&cli.BoolFlag{
		Name:        "veryverbose",
		Aliases:     []string{"vv"},
		Usage:       "Set logging level more verbose to include debug level logs",
		Destination: &Opts.VeryVerbose,
	},

	&cli.StringSliceFlag{
		Name:        "log-ids",
		Usage:       "Always emit logging for these ids, comma separated",
		Destination: &Opts.LogIDs,
	},
}

var Opts struct {
	Verbose     bool
	VeryVerbose bool
	LogIDs      cli.StringSlice
}

func Setup() {
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelWarn)
	if Opts.Verbose {
		logLevel.Set(slog.LevelInfo)
	}
	if Opts.VeryVerbose {
		logLevel.Set(slog.LevelDebug)
	}

	h := new(hlog.Handler)
	h = h.WithLevel(logLevel.Level())
	logIDs := Opts.LogIDs.Value()
	if len(logIDs) > 0 {
		for _, id := range logIDs {
			h = h.WithAttrLevel(slog.String("id", id), slog.LevelDebug)
		}
	}

	slog.SetDefault(slog.New(h))
}

var (
	Default = slog.Default
	Debug   = slog.Debug
	Info    = slog.Info
	Warn    = slog.Warn
	Error   = slog.Error
	With    = slog.With
)

func Dump(v any) {
	switch vt := v.(type) {
	case string:
		slog.Info(vt)
	default:
		slog.Info(utter.Sdump(v))
	}
}
