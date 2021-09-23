package main

// https://blog.kowalczyk.info/article/Jl3G/https-for-free-in-go.html
// To run:
// go run main.go
// Command-line options:
//   -production : enables HTTPS on port 443
//   -redirect-to-https : redirect HTTP to HTTTPS

import (
	"strings"
	"os"
	"path/filepath"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
	log "github.com/sirupsen/logrus"
)

const (
	htmlIndex = `<html><body>Welcome!</body></html>`
	httpPort  = ":80"
)

var (
	flgRedirectHTTPToHTTPS = true
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, htmlIndex)
}

func makeServerFromMux(mux *http.ServeMux) *http.Server {
	// set timeouts so that a slow or malicious client doesn't
	// hold resources forever
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
}

func makeHTTPServer() *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/", handleIndex)
	return makeServerFromMux(mux)

}

func makeHTTPToHTTPSRedirectServer() *http.Server {
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		newURI := "https://" + r.Host + r.URL.String()
		http.Redirect(w, r, newURI, http.StatusFound)
	}
	mux := &http.ServeMux{}
	mux.HandleFunc("/", handleRedirect)
	return makeServerFromMux(mux)
}

func parseFlags() {
	// flag.BoolVar(&flgRedirectHTTPToHTTPS, "redirect-to-https", false, "if true, we redirect HTTP to HTTPS")
	flag.Parse()
}

func main() {
	parseFlags()
	var m *autocert.Manager

	wwwDir := filepath.Join("/", "www")

	var httpsSrv *http.Server
	{
		hostPolicy := func(ctx context.Context, host string) error {
			if !strings.Contains(host, ".") {
				return fmt.Errorf("Host is not valid: %s", host)
			}
			hostDir := filepath.Join(wwwDir, host)
			stat, err := os.Stat(hostDir)
			if err != nil {
				return fmt.Errorf("Host path does not exist: %s", hostDir)
			}
			if !stat.IsDir() {
				return fmt.Errorf("Host path %s is instead a file?", hostDir)
			}
			log.Infof("Allowing cert for %s", host)
			return nil
		}

		dataDir := filepath.Join(wwwDir, "certs")
		err := os.MkdirAll(dataDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		m = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: hostPolicy,
			Cache:      autocert.DirCache(dataDir),
		}

		httpsSrv = makeHTTPServer()
		httpsSrv.Addr = ":443"
		httpsSrv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

		go func() {
			fmt.Printf("Starting HTTPS server on %s\n", httpsSrv.Addr)
			err := httpsSrv.ListenAndServeTLS("", "")
			if err != nil {
				log.Fatalf("httpsSrv.ListendAndServeTLS() failed with %s", err)
			}
		}()
	}

	var httpSrv *http.Server
	if flgRedirectHTTPToHTTPS {
		httpSrv = makeHTTPToHTTPSRedirectServer()
	} else {
		httpSrv = makeHTTPServer()
	}
	// allow autocert handle Let's Encrypt callbacks over http
	if m != nil {
		httpSrv.Handler = m.HTTPHandler(httpSrv.Handler)
	}

	httpSrv.Addr = httpPort
	fmt.Printf("Starting HTTP server on %s\n", httpPort)
	err := httpSrv.ListenAndServe()
	if err != nil {
		log.Fatalf("httpSrv.ListenAndServe() failed with %s", err)
	}
}
