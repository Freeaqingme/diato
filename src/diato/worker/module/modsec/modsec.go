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
package modsec

import (
	"diato/worker"
	"fmt"
	"net/http"

	"go-modsecurity"
	"io/ioutil"
	"log"
	"bytes"
	"runtime"
)

const name = "modsec"

func init() {
	worker.RegisterModule(newModule)
}

type module struct {
	modsec  *modsecurity.Modsecurity
	ruleset *modsecurity.RuleSet
}

func newModule(w *worker.Worker) ([]worker.Module, error) {
	modsec, err := modsecurity.NewModsecurity()
	if err != nil {
		return nil, err
	}

	ruleset := modsec.NewRuleSet()
	//fmt.Println(ruleset.AddFile("/home/dolf/Projects/diato/modsec.conf"))
	rules := "SecRuleEngine On\n" +
		"SecRequestBodyAccess On\n" +
		"SecRequestBodyLimit 102400\n" +
		"SecDebugLog /dev/stderr\n" +
		"SecDebugLogLevel 9\n" +
		"SecRule REQUEST_URI|ARGS|REQUEST_BODY \"usernaaam\" \"id:1,phase:2,log,deny,msg:'Access Denied'\"\n" +
	"SecRule REQUEST_BODY \"usernaaam\" \"id:3,phase:2,deny\"\n" +
	"SecRule REQUEST_BODY \"usernaaam\" \"phase:2, t:none, deny,msg:'Matched some_bad_string', status:500,auditlog, id:3333\"\n" +
		"SecRule ARGS \"@streq test\" \"id:2,phase:2,deny\"\n"
	fmt.Println("rule errors", ruleset.AddRules(rules))

	return []worker.Module{&module{
		modsec:  modsec,
		ruleset: ruleset,
	}}, nil
}

func (m *module) Enabled() bool {
	return true
}

func (m *module) Name() string {
	return name
}

func (m *module) ProcessRequest(req *http.Request) {
	fmt.Println("sending request to modsec...")

	// There is no way to get the local addr from the http.Request object?
	txn, _ := m.ruleset.NewTransaction(req.RemoteAddr, "127.0.0.1:80")

	url := req.URL      // TODO: Check if it works with https
	url.Host = req.Host // req.URL.host seems to be always empty at this stage, so we set it
	httpVersion := fmt.Sprintf("%d.%d", req.ProtoMajor, req.ProtoMinor)

	txn.ProcessUri(url.String(), req.Method, httpVersion)
	// We don't currently parse request headers as it appears to corrupt our memory, somewhere
	//for key,values := range req.Header{
		//for _, value := range values {
			//txn.AddRequestHeader(key, value)
		//}
	//}

	txn.ProcessRequestHeaders()

	fmt.Println(txn.ShouldIntervene())


	fmt.Println("Now scanning body")
	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		if err != nil {
			log.Printf("Error reading body: %v", err)
			return
		}

		fmt.Println("body:", string(body))
		fmt.Println(txn.AppendRequestBody(body))
		fmt.Println(txn.ProcessRequestBody())
	}

	fmt.Println(txn.ShouldIntervene())
	txn.Cleanup()
	runtime.GC()
}
