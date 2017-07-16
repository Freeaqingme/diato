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
package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"diato/util/stop"

	"github.com/Freeaqingme/go-proxyproto"
)

type httpBind struct {
	name       string
	listen     string
	proxyProto bool
	hasSsl     bool
}

func (s *Server) Listen(bind *httpBind) error {
	ln, err := net.Listen("tcp", bind.listen)
	if err != nil {
		return err
	}

	logMsgSuffix := []string{}
	if bind.proxyProto {
		ln = &proxyproto.Listener{Listener: ln}
		logMsgSuffix = append(logMsgSuffix, "Proxy Protocol")
	}

	if bind.hasSsl {
		ln, err = s.tlsListen(ln)
		if err != nil {
			return err
		}
		logMsgSuffix = append(logMsgSuffix, "TLS")
	}

	if len(logMsgSuffix) == 0 {
		log.Printf("Now listening on %s: %s\n", bind.name, bind.listen)
	} else {
		log.Printf("Now listening on %s: %s (%s)\n",
			bind.name, bind.listen, strings.Join(logMsgSuffix, ", "))
	}

	stopper := stop.NewStopper(func() {
		ln.Close()
	})

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if stopper.IsStopping() {
					return
				}
				panic(err.Error())
			}
			go s.handleConn(bind, conn)
		}
	}()

	return nil
}

func (s *Server) handleConn(bind *httpBind, conn net.Conn) {
	path := s.httpSocketPath
	if bind.hasSsl {
		path = s.httpsSocketPath
	}

	client, err := net.Dial("unix", path)
	if err != nil {
		log.Fatalf("Dial failed: %v", err)
	}

	hdr, err := s.getProxyProtoHeader(conn)
	if err != nil {
		panic("TODO, error handling: " + err.Error())
	}
	_, err = fmt.Fprint(client, hdr)
	if err != nil {
		panic("Todo: error handling")
	}

	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(client, conn)
	}()
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(conn, client)
	}()
}

// Originally derived from https://github.com/nabeken/mikoi
// Released under BSD-3 license, by Tanabe Ken-ichi
func (s *Server) getProxyProtoHeader(conn net.Conn) (string, error) {
	saddr, sport, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return "", err
	}

	daddr, dport, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return "", err
	}

	raddr, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		return "", errors.New("Cannot proxy protocol other than TCP4 or TCP6")
	}

	var tcpStr string
	if rip4 := raddr.IP.To4(); len(rip4) == net.IPv4len {
		tcpStr = "TCP4"
	} else if len(raddr.IP) == net.IPv6len {
		tcpStr = "TCP6"
	} else {
		return "", errors.New("Unrecognized protocol type")
	}

	hdr := fmt.Sprintf("PROXY %s %s %s %s %s\r\n", tcpStr, saddr, daddr, sport, dport)
	return hdr, err
}
