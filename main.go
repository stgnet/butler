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
	"time"

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

// ### TODO ### turn this into a cache
func validSHost(host string) bool {
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
			log.Infof("validSHost: Not HTTPS: Resolved '%s' as %v", host, ip.String())
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
			log.Infof("Serving request: %s", r.Host+r.URL.String())
			// w.WriteHeader(http.StatusNotFound)
			errorHandler(w, r, http.StatusNotFound,
				fmt.Errorf("unimplemented //%s/%s", r.Host, r.URL.String()))
			/*
				html, err := json.MarshalIndent(r, "", "  ")
				if err != nil {
					log.Errorf("Failed json: %v", err)
				}
			*/
		}
	}
	mux := &http.ServeMux{}
	mux.HandleFunc("/", handleRedirect)
	return makeServerFromMux(mux)
}

func dirExists(path string) (bool, bool) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, false
	}
	if stat.IsDir() {
		return true, true
	}
	return false, true
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.WriteHeader(status)
	/*
	   if status == http.StatusNotFound {
	       fmt.Fprint(w, "custom 404")
	   }
	*/
	fmt.Fprintf(w, "Error %d: %v", status, err)
}

func main() {

	if len(os.Args[1:]) > 0 {
		locate(os.Stdout, filepath.Join(os.Args[1:]...))
		return
	}
	var m *autocert.Manager

	wwwDir := filepath.Join("/", "www")
	dir, _ := dirExists(wwwDir)
	if !dir {
		wwwDir = filepath.Join(".", "www")
		dir, _ = dirExists(wwwDir)
		if !dir {
			// TODO: assume operation out of cwd and show index
			log.Fatalf("Unable to locate www directory")
		}
	}

	var httpsSrv *http.Server
	{
		hostPolicy := func(ctx context.Context, host string) error {
			if !validHost(host) {
				return fmt.Errorf("Host is not valid: %s", host)
			}
			hostDir := filepath.Join(wwwDir, host)
			dir, exists := dirExists(hostDir)
			if !exists {
				return fmt.Errorf("Host path does not exist: %s", hostDir)
			}
			if !dir {
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
