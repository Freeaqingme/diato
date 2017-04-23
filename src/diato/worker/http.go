package worker

import (
	"context"
	"diato/util/stop"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
	"fmt"
	"io"
	//"bytes"
)

func (w *Worker) httpListen() {
	srv := &http.Server{
		ReadTimeout: 5 * time.Second,
		IdleTimeout: 120 * time.Second,
		Handler:     w.newHttpHandler(),
	}

	stop.NewStopper(func() {
		timeout := 1 * time.Minute
		log.Printf("Gracefully stopping HTTP Server. This can take a while (%s)", timeout)
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Minute)
		srv.Shutdown(ctx)
		log.Print("HTTP Server gracefully stopped on Worker side")
	})

	srv.Serve(w.ipcSocket)
}



func (w *Worker) newHttpHandler() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = "127.0.0.1:8080"

		log.Printf("%s '%s %s %s' -> %s '%s'\n",
			req.RemoteAddr,
			req.Method, req.RequestURI,
			req.Host,
			req.URL.Host,
			req.UserAgent(),
		)
	}

	return &httputil.ReverseProxy{
		Director:      director,
		FlushInterval: 10 * time.Millisecond,
		Transport: &httpTransport{
			&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				Dial: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				MaxIdleConnsPerHost: 64,
			},
		},
	}
}

type httpTransport struct {
	http.RoundTripper
}

func (t *httpTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.RoundTripper.RoundTrip(req)

	fmt.Println(resp.Status)
	for k, v := range resp.Header {
		fmt.Println(k,v)
	}

	r := resp.Body

	var writer *io.PipeWriter
	resp.Body, writer = io.Pipe()

	go func() {
		defer resp.Body.Close()
		buf := make([]byte, 0, 4096)
		for {
			n, err := r.Read(buf[:cap(buf)])
			buf = buf[:n]
			if n == 0 {
				if err == nil {
					continue
				}
				if err == io.EOF {
					break
				}
				return
			}

			fmt.Print(string(buf)) // processing here?

			writer.Write(buf)
			if err != nil && err != io.EOF {
				return
			}
		}
	}()
	return resp, nil
}
