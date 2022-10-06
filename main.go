package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/jamescun/httplog/httplog"
)

// Version is the semantic release version of this build of httplog.
var Version = "0.0.0"

var (
	listenAddr   = flag.String("listen", "localhost:8080", "configure the listening address for the HTTP server")
	responseBody = flag.String("response", "", "configure the canned body sent in response to all requests")
	responseCode = flag.Int("response-code", 200, "configure the HTTP status code sent in response requests")
	logJSON      = flag.Bool("json", false, "log all requests as JSON rather than human readable text")
)

const Usage = `httplog v%s

httplog is a command line tool that launches a local HTTP server that logs all
requests it receives, replying with a canned response.

Usage: httplog [options...]

Options:
  --help                        show helpful information
  --listen         <host:port>  configure the listening address for the HTTP
                                server (default localhost:8080)
  --response       <text>       configure the canned body sent in response to
                                all requests (default none)
  --response-code  <code>       configure the HTTP status code sent in response
                                to all requests (default 200)
  --json                        log all requests as JSON rather than human
                                readable text
`

func main() {
	flag.Usage = func() { fmt.Fprintf(os.Stderr, Usage, Version) }
	flag.Parse()

	srv := httplog.NewServer(128)

	srv.ResponseCode = *responseCode

	if *responseBody != "" {
		srv.ResponseBody = []byte(*responseBody)
	}

	if *logJSON {
		go httplog.JSONLogger(os.Stdout, srv.Requests())
	} else {
		go httplog.TextLogger(os.Stdout, srv.Requests())
	}

	s := &http.Server{
		Addr:    *listenAddr,
		Handler: srv,
	}

	fmt.Fprintf(os.Stderr, "server listening on %s...\n", *listenAddr)

	if err := s.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(1)
	}
}
