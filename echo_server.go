package main

import (
	"log"
	"net/http"
	"strconv"
)

// EchoServer is http server, it echos all requests received.
type EchoServer struct {
	port int
	quit chan struct{}
}

// NewEchoServer returns a new EchoServer.
func NewEchoServer(port int) *EchoServer {
	return &EchoServer{port: port, quit: make(chan struct{})}
}

// Start runs the server, and starts handling requests.
func (s *EchoServer) Start() error {
	errChan := make(chan error)
	log.Printf("echo server is serving at :%d", s.port)
	go func() {
		select {
		case errChan <- http.ListenAndServe(":"+strconv.Itoa(s.port), s):
		case <-s.quit:
			errChan <- nil

		}
	}()
	return <-errChan
}

// Stop shuts the server down.
func (s *EchoServer) Stop() {
	close(s.quit)
	log.Printf("echo server at :%d stopped", s.port)
}

// ServerHTTP implements http.Handler.
func (s *EchoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("request received at :%d", s.port)
	w.Write([]byte("request from " + r.RemoteAddr + " received, handled as " + "http://" + r.Host + r.RequestURI))
}

// Addr returns url address of the server.
func (s *EchoServer) Addr() string {
	return "http://localhost:" + strconv.Itoa(s.port) + "/abc?q=abc"
}
