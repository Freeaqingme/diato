// Diato - Reverse Proxying for Hipsters
//
// Copyright 2016-2017 Dolf Schimmel
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package worker

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "diato/pb"
	"diato/util/stop"

	"github.com/Freeaqingme/go-proxyproto"
)

func (w *Worker) httpGetListener(tls bool) (net.Listener, error) {
	var fd uintptr
	if tls {
		fd = 5
	} else {
		fd = 4
	}

	httpLn, err := net.FileListener(os.NewFile(fd, "[socket]"))
	if err != nil {
		return nil, err
	}

	return &proxyproto.Listener{Listener: httpLn}, nil
}

func (w *Worker) httpListen(httpSocket net.Listener, tls bool) {
	srv := &http.Server{
		ReadTimeout: 5 * time.Second,
		IdleTimeout: 120 * time.Second,
		Handler:     w.newHttpHandler(tls),
	}

	stop.NewStopper(func() {
		timeout := 1 * time.Minute
		log.Printf("Gracefully stopping HTTP Server. This can take a while (%s)", timeout)
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Minute)
		srv.Shutdown(ctx)
		log.Print("HTTP Server gracefully stopped on Worker side")
	})

	srv.Serve(httpSocket)
}

func (w *Worker) newHttpHandler(tls bool) *ReverseProxy {
	director := func(req *http.Request) {
		w.modules.ProcessRequest(req)
		// local addr: fmt.Println(req.Context().Value(http.LocalAddrContextKey).(net.Addr))

		var err error
		ctxInfo := req.Context().Value("diato").(*ContextInfo)
		req.URL.Scheme = "http"
		req.URL.Host, err = w.getHttpBackend(req)
		if err != nil {
			log.Printf("Couild not determine backend for client %s: %s", req.RemoteAddr, err.Error())
			req.URL.Host = "" // TODO: This is suboptimal as modsec will crash upon an empty url
			return
		}
		if tls {
			req.Header.Add("X-Forwarded-Proto", "https")
		}

		log.Printf("%s %s '%s %s %s' -> %s '%s' (SSL: %t)\n",
			req.RemoteAddr,
			ctxInfo.RequestIdString(),
			req.Method, req.RequestURI,
			req.Host,
			req.URL.Host,
			req.UserAgent(),
			tls,
		)
	}

	return &ReverseProxy{
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
		ModifyResponse: func(r *http.Response) error {
			// TODO: If backend is unavailable this header is never added
			r.Header.Add("X-Powered-By", "Diato")

			go w.modules.PostModifyResponse(r)
			return nil
		},
	}
}

func (w *Worker) getHttpBackend(req *http.Request) (string, error) {
	r, err := w.userBackend.GetServerForUser(
		req.Context(),
		&pb.UserBackendRequest{Name: req.Host},
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", r.Server, r.Port), nil
}

type httpTransport struct {
	http.RoundTripper
}

func (t *httpTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	//fmt.Println(resp.Status)
	//for k, v := range resp.Header {
	//fmt.Println(k, v)
	//}

	r := resp.Body

	var writer *io.PipeWriter
	resp.Body, writer = io.Pipe()

	go func() {
		defer writer.Close()
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

			//fmt.Print(string(buf)) // processing here?

			writer.Write(buf)
			if err != nil && err != io.EOF {
				return
			}
		}
	}()
	return resp, nil
}
