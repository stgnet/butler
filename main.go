package main

// https://blog.kowalczyk.info/article/Jl3G/https-for-free-in-go.html

import (
	"context"
	"crypto/tls"
	// "encoding/json"
	// "bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	// "strings"
	// "io/ioutil"
	"time"
	// "errors"
	// "net/url"

	// "github.com/goccy/go-yaml"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

const (
	htmlIndex = `<html><body>Welcome!</body></html>`
	httpPort  = ":80"
)

var (
	flgRedirectHTTPToHTTPS = true
	root                   Tray
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// entry point for all butler handled requests
	// hostDir := filepath.Join(wwwDir, r.Host, r.URL.Path)

	// first locate our virtual host
	host := r.Host
	if host == "" {
		host = "defaut"
	}
	var tray Tray
	var tErr error
	if host == "localhost" {
		tray = root
	} else {
		tray, tErr = root.Pass(host)
		if tErr != nil {
			errorHandler(w, r, Broke{Status: http.StatusNotFound, Problem: tErr})
			return
		}
	}

	file := "index"
	paths := splitpath(r.URL.Path)
	last := len(paths) - 1
	for index, path := range paths {
		next, nErr := tray.Pass(path)
		if nErr != nil {
			if index == last {
				// presume the last entry is the file
				file = path
				break
			} else {
				errorHandler(w, r, Broke{Status: http.StatusNotFound, Problem: nErr})
				return
			}
		}
		tray = next
	}

	glass, gErr := tray.Glass(file)
	if gErr != nil {
		errorHandler(w, r, Broke{Status: http.StatusNotFound, Problem: gErr})
		return
	}

	w.Header().Set("Content-Type", glass.Type())
	glass.Pour(w)
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
	mux.HandleFunc("/", handleRequest)
	return makeServerFromMux(mux)

}

// ### TODO ### turn this into a cache
func validSHost(host string) bool {
	if len(host) == 0 {
		return false
	}
	if host[0] >= '0' && host[0] <= '9' {
		return false
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		log.Infof("validSHost: Not HTTPS: Could not resolve '%s': %v", host, err)
		return false
	}
	if len(ips) == 0 {
		log.Infof("validSHost: Not HTTPS: no addresses returned for lookup '%s' - no HTTPS", host)
		return false
	}
	for _, ip := range ips {
		if ip.IsPrivate() || ip.IsLoopback() {
			// log.Infof("validSHost: Not HTTPS: Resolved '%s' as %v", host, ip.String())
			return false
		}
		// another step: if ip matches interface addr, return true
	}
	return true
}
func validHost(host string) bool {
	if host[0] >= '0' && host[0] <= '9' {
		return false
	}
	// "xyz" is actually a valid hostname
	// return strings.Contains(host, ".")
	return true
}

func makeHTTPToHTTPSRedirectServer() *http.Server {
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		if validSHost(r.Host) {
			newURI := "https://" + r.Host + r.URL.String()
			log.Infof("Sending redirect from %s from %s", r.Host+r.URL.String(), newURI)
			http.Redirect(w, r, newURI, http.StatusFound)
		} else {
			// log.Infof("Serving HTTP request: %s", r.Host+r.URL.String())
			handleRequest(w, r)
			return
		}
	}
	mux := &http.ServeMux{}
	mux.HandleFunc("/", handleRedirect)
	return makeServerFromMux(mux)
}

func dirExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return true
	}
	return false
}

type Broke struct {
	Status  int
	Problem error
}

func (e Broke) Error() string {
	return fmt.Sprintf("Error %d %v", e.Status, e.Problem)
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	switch e := err.(type) {
	case Broke:
		w.WriteHeader(e.Status)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	io.WriteString(w, err.Error())
	log.Infof("%s%v: %v", r.Host, r.URL, err)
}

var roots = []string{
	"/www",
	"/var/www",
	"./www",
	".",
}

func main() {
	if len(os.Args[1:]) > 0 {
		locate(os.Stdout, filepath.Join(os.Args[1:]...))
		return
	}

	lifetime := time.NewTimer(time.Hour * 6)
	go func() {
		<-lifetime.C
		log.Infof("Shutting down on lifetime")
		os.Exit(1)
	}()

	for _, path := range roots {
		if !dirExists(path) {
			continue
		}
		root = Cast(path, nil)
		break
	}
	if root == nil {
		panic("No directory found for root tray")
	}
	log.Infof("Using %s for root tray", root.Path())

	var m *autocert.Manager

	var httpsSrv *http.Server
	{
		hostPolicy := func(ctx context.Context, host string) error {
			if !validHost(host) {
				return fmt.Errorf("Host is not valid: %s", host)
			}
			// this was checking to make sure host was valid dir
			log.Infof("Allowing cert for %s", host)
			return nil
		}

		dataDir := filepath.Join(root.Path(), "certs")
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
