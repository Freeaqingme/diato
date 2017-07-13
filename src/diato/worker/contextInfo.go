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
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/Freeaqingme/publicsuffix-go/publicsuffix"
	"github.com/bwmarrin/snowflake"
	ua "github.com/mssola/user_agent"
)

var snowflakeGenerator *snowflake.Node

func init() {
	rand.Seed(time.Now().UnixNano())
	hostId := rand.Uint64() % 1024

	var err error
	snowflakeGenerator, err = snowflake.NewNode(int64(hostId))
	if err != nil {
		panic(err.Error())
	}
}

type ContextInfo struct {
	timeStart time.Time

	// Keep it at 64 bits because modsecurity uses 64 bit id's for logging
	requestId int64
	sld       string
	userAgent *ua.UserAgent
}

func getRequestWithContextInfo(r *http.Request) *http.Request {
	contextInfo := &ContextInfo{
		timeStart: time.Now(),
		requestId: snowflakeGenerator.Generate().Int64(),
		userAgent: ua.New(r.UserAgent()),
	}
	contextInfo.setSld(r)

	ctx := r.Context()
	return r.WithContext(context.WithValue(ctx, "diato", contextInfo))
}

func (i *ContextInfo) TimeStart() time.Time {
	return i.timeStart
}

func (i *ContextInfo) RequestId() int64 {
	return i.requestId
}

func (i *ContextInfo) RequestIdString() string {
	return strconv.FormatInt(i.RequestId(), 36)
}

func (i *ContextInfo) Sld() string {
	return i.sld
}

func (i *ContextInfo) UserAgent() *ua.UserAgent {
	return i.userAgent
}

func (i *ContextInfo) setSld(r *http.Request) {
	i.sld, _ = publicsuffix.DomainFromListWithOptions(
		publicsuffix.DefaultList,
		r.Host,
		&publicsuffix.FindOptions{IgnorePrivate: true},
	)

	if i.sld == "" {
		i.sld = r.Host
	}
}
