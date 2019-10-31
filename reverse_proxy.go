package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

// ReverseProxy proxies requests to given hosts by order.
type ReverseProxy struct {
	hosts []*url.URL
	port  int
	quit  chan struct{}

	nextHostIndex int
}

// NewReverseProxy returns a new ReverseProxy.
func NewReverseProxy(hosts []string, port int) (*ReverseProxy, error) {
	urls := make([]*url.URL, 0, len(hosts))
	for _, h := range hosts {
		u, err := url.Parse(h)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", h, err)
		}
		urls = append(urls, u)
	}
	return &ReverseProxy{hosts: urls, port: port, quit: make(chan struct{})}, nil
}

// Start runs the ReverseProxy, and starts proxying requests.
func (r *ReverseProxy) Start() error {
	errChan := make(chan error)
	log.Printf("reverse proxy is serving at :%d", r.port)
	go func() {
		select {
		case errChan <- http.ListenAndServe(":"+strconv.Itoa(r.port), r):
		case <-r.quit:
			errChan <- nil
		}
	}()
	return <-errChan
}

// Stop shuts ReverseProxy down.
func (r *ReverseProxy) Stop() {
	close(r.quit)
	log.Printf("reverse proxy at :%d stopped", r.port)
}

// ServeHTTP implements http.Handler
func (r *ReverseProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("request received at :%d", r.port)
	host := r.hosts[r.nextHostIndex]
	r.nextHostIndex = (r.nextHostIndex + 1) % len(r.hosts)
	// proxy := httputil.NewSingleHostReverseProxy(host)
	// httputil.NewSingleHostReverseProxy doesn't overwrite the host of request
	proxy := r.proxyOf(host)
	proxy.ServeHTTP(w, req)
}

func (r *ReverseProxy) proxyOf(target *url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme // no effect
			req.URL.Host = target.Host     // no effect
			req.URL.Path = r.singleJoiningSlash(target.Path, req.URL.Path)
			if req.URL.RawQuery == "" || target.RawQuery == "" {
				req.URL.RawQuery = target.RawQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = target.RawQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				req.Header.Set("User-Agent", "")
			}
			req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
			req.Host = target.Host
		},
	}
}

func (r *ReverseProxy) singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	default:
		return a + b
	}
}
