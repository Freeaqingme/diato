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

package v1

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"diato/worker"

	"github.com/blang/semver"
)

func (m *mapping) getStructuredRequest(req *http.Request, resp *http.Response) *httpRequest {
	ctx := req.Context()
	ctxInfo := ctx.Value("diato").(*worker.ContextInfo)

	hostname, _ := os.Hostname() // TODO: Determine this once at startup and then use that?
	remoteAddr, _, _ := net.SplitHostPort(req.RemoteAddr)
	//fmt.Println(ctx.Value(http.LocalAddrContextKey).(*net.UnixAddr)) // ./http.socket :(

	var url string
	if len(req.URL.RawQuery) == 0 {
		url = req.URL.Path
	} else {
		url = fmt.Sprintf("%s?%s", req.URL.Path, req.URL.RawQuery)
	}

	return &httpRequest{
		Host:      req.Host,
		Url:       url,
		Path:      req.URL.Path,
		Query:     req.URL.RawQuery,
		Timestamp: ctxInfo.TimeStart(),
		Duration:  time.Now().Sub(ctxInfo.TimeStart()).Seconds(),
		SLD:       ctxInfo.Sld(),

		RemoteIp: remoteAddr,
		//Local_ip:      localAddr,
		ResponseCode: resp.StatusCode,
		Method:       req.Method,
		// TODO: Geoip
		HttpVersion: float32(req.ProtoMajor) + (float32(req.ProtoMinor) / 10),
		Referrer:    req.Referer(),
		UserAgent:   m.getUserAgent(req),

		Diato: diatoInfo{
			Hostname: hostname,
		},
	}
}

func (m *mapping) getUserAgent(req *http.Request) userAgent {
	ctx := req.Context()
	ctxInfo := ctx.Value("diato").(*worker.ContextInfo)
	ua := ctxInfo.UserAgent()

	browserName, browserVersion := ua.Browser()
	browserEngineName, browserEngineVersion := ua.Engine()

	var browserMajor, browserMinor uint64
	browserSemver, err := semver.Make(browserVersion)
	if err == nil {
		browserMajor = browserSemver.Major
		browserMinor = browserSemver.Minor
	}

	var engineMajor, engineMinor uint64
	engineSemver, err := semver.Make(browserEngineVersion)
	if err == nil {
		engineMajor = engineSemver.Major
		engineMinor = engineSemver.Minor
	}

	return userAgent{
		Raw: req.UserAgent(),
		Browser: browser{
			Name:         browserName,
			Version:      browserVersion,
			VersionMajor: int(browserMajor),
			VersionMinor: int(browserMinor),
			Engine: browserEngine{
				Name:         browserEngineName,
				Version:      browserEngineVersion,
				VersionMajor: int(engineMajor),
				VersionMinor: int(engineMinor),
			},
		},
		Platform: ua.Platform(),
		Os:       ua.OS(),
		Mobile:   ua.Mobile(),
		Bot:      ua.Bot(),
	}
}
