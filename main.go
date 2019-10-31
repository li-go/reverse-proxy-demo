package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"time"
)

func main() {
	var hosts []string

	svr := runEchoServer(18001)
	defer svr.Stop()
	hosts = append(hosts, svr.Addr())

	svr = runEchoServer(18002)
	defer svr.Stop()
	hosts = append(hosts, svr.Addr())

	svr = runEchoServer(18003)
	defer svr.Stop()
	hosts = append(hosts, svr.Addr())

	port := 18000
	proxy, err := NewReverseProxy(hosts, port)
	if err != nil {
		log.Printf("failed to initialize reverse proxy at :%d: %v", port, err)
		return
	}
	go func() {
		err := proxy.Start()
		if err != nil {
			log.Printf("failed to start reverse proxy at :%d: %v", port, err)
		}
	}()

	// dummy requests after proxy server started
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second)
			res, _ := http.Get("http://localhost:18000/hello?q=hello")
			bs, _ := httputil.DumpResponse(res, true)
			log.Print("\n" + string(bs))
		}
	}()

	// wait for interrupt signal to shut down
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt)
	select {
	case <-signChan:
	}
	time.Sleep(50 * time.Millisecond)
	log.Print("done!")
}

func runEchoServer(port int) *EchoServer {
	svr := NewEchoServer(port)
	go func() {
		err := svr.Start()
		if err != nil {
			log.Printf("failed to start server at :%d: %v", port, err)
		}
	}()
	return svr
}
