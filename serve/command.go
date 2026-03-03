package serve

import (
	"fmt"
	"net"
	"net/http"

	"github.com/iand/genster/logging"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:   "serve",
	Usage:  "Serve the built site from the pub directory",
	Action: serveAction,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "pub",
			Aliases:     []string{"p"},
			Usage:       "Path to the pub directory to serve",
			Required:    true,
			Destination: &serveOpts.pubDir,
		},
		&cli.StringFlag{
			Name:        "addr",
			Aliases:     []string{"a"},
			Usage:       "Address to listen on",
			Value:       "localhost:1313",
			Destination: &serveOpts.addr,
		},
	}, logging.Flags...),
}

var serveOpts struct {
	pubDir string
	addr   string
}

func serveAction(cc *cli.Context) error {
	logging.Setup()

	ln, err := net.Listen("tcp", serveOpts.addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", serveOpts.addr, err)
	}

	fmt.Printf("Serving %s\n", serveOpts.pubDir)
	fmt.Printf("Listening on http://%s/\n", ln.Addr())

	return http.Serve(ln, http.FileServer(http.Dir(serveOpts.pubDir)))
}
