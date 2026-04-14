package build

import (
	"context"
	"fmt"

	"github.com/iand/genster/logging"
	"github.com/urfave/cli/v3"
)

var Command = &cli.Command{
	Name:   "build",
	Usage:  "Render a content directory to a static HTML site",
	Action: buildAction,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "input",
			Aliases:     []string{"i"},
			Usage:       "Path to the input directory",
			Required:    true,
			Destination: &buildOpts.contentDir,
		},
		&cli.StringFlag{
			Name:        "pub",
			Aliases:     []string{"p"},
			Usage:       "Path to the output pub directory",
			Required:    true,
			Destination: &buildOpts.pubDir,
		},
		&cli.StringFlag{
			Name:        "assets",
			Aliases:     []string{"a"},
			Usage:       "Path to static assets directory (CSS, JS); embedded defaults used if not set",
			Destination: &buildOpts.assetsDir,
		},
		&cli.StringFlag{
			Name:        "base-url",
			Usage:       "Base URL of the site (e.g. https://example.com) used to build absolute URLs in sitemap.xml; sitemap is omitted if not set",
			Destination: &buildOpts.baseURL,
		},
		&cli.BoolFlag{
			Name:        "include-drafts",
			Usage:       "Include pages marked draft: true in the output",
			Destination: &buildOpts.includeDrafts,
		},
		&cli.BoolFlag{
			Name:        "include-private",
			Usage:       "Include body content of pages marked private: yes in the output",
			Destination: &buildOpts.includePrivate,
		},
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "Add a debug footer to every rendered page",
			Destination: &buildOpts.debug,
		},
	}, logging.Flags...),
}

var buildOpts struct {
	contentDir     string
	pubDir         string
	assetsDir      string
	baseURL        string
	includeDrafts  bool
	includePrivate bool
	debug          bool
}

func buildAction(ctx context.Context, cc *cli.Command) error {
	logging.Setup()

	b := &Builder{
		ContentDir:     buildOpts.contentDir,
		PubDir:         buildOpts.pubDir,
		AssetsDir:      buildOpts.assetsDir,
		BaseURL:        buildOpts.baseURL,
		IncludeDrafts:  buildOpts.includeDrafts,
		IncludePrivate: buildOpts.includePrivate,
		Debug:          buildOpts.debug,
	}

	if err := b.Build(); err != nil {
		return err
	}

	fmt.Printf("Site built in %s\n", buildOpts.pubDir)
	return nil
}
