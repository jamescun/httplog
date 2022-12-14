package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/jamescun/httplog/httplog"
)

// Version is the semantic release version of this build of httplog.
var Version = "0.0.0"

var (
	listenAddr   = flag.String("listen", "localhost:8080", "configure the listening address for the HTTP server")
	responseBody = flag.String("response", "", "configure the canned body sent in response to all requests")
	responseCode = flag.Int("response-code", 200, "configure the HTTP status code sent in response requests")
	logJSON      = flag.Bool("json", false, "log all requests as JSON rather than human readable text")
	tlsSelfCert  = flag.Bool("tls-self-cert", false, "enable TLS with a self-signed certificate")
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
  --tls-self-cert               enable TLS with a self-signed certificate
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

	if *tlsSelfCert {
		cfg, err := generateSelfCertConfig()
		if err != nil {
			exitError(1, "could not generate self signed certificate: %s", err)
		}

		s.TLSConfig = cfg
	}

	fmt.Fprintf(os.Stderr, "server listening on %s...\n", *listenAddr)

	if s.TLSConfig != nil {
		if err := s.ListenAndServeTLS("", ""); err != nil {
			exitError(1, "server: %s", err)
		}
	} else {
		if err := s.ListenAndServe(); err != nil {
			exitError(1, "server: %s", err)
		}
	}
}

func exitError(code int, format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func generateSelfCertConfig() (*tls.Config, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}

	cert := &x509.Certificate{
		Version:      3,
		SerialNumber: big.NewInt(1),
		Issuer:       pkix.Name{CommonName: "httplog"},
		Subject:      pkix.Name{CommonName: "httplog"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(720 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"httplog"},
	}

	bytes, err := x509.CreateCertificate(rand.Reader, cert, cert, privateKey.Public(), privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{bytes},
				PrivateKey:  privateKey,
			},
		},
	}

	return cfg, nil
}
