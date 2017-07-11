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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"diato/config"
	"diato/module/modsec/pb"
	"diato/worker"

	"github.com/golang/protobuf/ptypes/empty"
	"go-modsecurity"
)

const name = "modsec"

func init() {
	worker.RegisterModule(newModule)
}

type module struct {
	*worker.ModuleBase

	enabled bool
	modsec  *modsecurity.Modsecurity
	ruleset *modsecurity.RuleSet

	worker *worker.Worker
	grpc   pb.ModuleModsecClient
}

func newModule(w *worker.Worker, config *config.Config) ([]worker.Module, error) {
	if !config.Modsec.Enabled {
		return []worker.Module{&module{enabled: false}}, nil
	}

	modsec, err := modsecurity.NewModsecurity()
	if err != nil {
		return nil, err
	}

	modsec.SetServerLogCallback(func(msg string) {
		log.Println(msg)
	})

	log.Printf("Initialized libmodsecurity: %s", modsec.WhoAmI())

	module := &module{
		enabled: true,
		worker:  w,
		modsec:  modsec,
	}

	if err := module.loadRules(); err != nil {
		return nil, err
	}

	return []worker.Module{module}, nil

}

func (m *module) Enabled() bool {
	return m.enabled
}

func (m *module) Name() string {
	return name
}

func (m *module) ProcessRequest(req *http.Request) {
	// There is no way to get the local addr from the http.Request object?
	txn, _ := m.ruleset.NewTransaction(req.RemoteAddr, "127.0.0.1:80")
	defer func() {
		// TODO: If we also start processing responses,
		// this should only be executed afterwards
		txn.ProcessLogging()
		txn.Cleanup()
	}()

	url := req.URL // TODO: Check if it works with https
	//url.Host = req.Host // req.URL.host seems to be always empty at this stage, so we set it
	httpVersion := fmt.Sprintf("%d.%d", req.ProtoMajor, req.ProtoMinor)

	txn.ProcessUri(url.String(), req.Method, httpVersion)

	// TODO: An occasional fatal error: concurrent map iteration and map write
	//		 but is it really, or is our memory simply somewhere corrupted?
	for key, values := range req.Header {
		for _, value := range values {
			txn.AddRequestHeader([]byte(key), []byte(value))
		}
	}

	txn.ProcessRequestHeaders()

	if req.Body != nil {
		if txn.ShouldIntervene() {
			log.Printf("Should intervene in request from %s for %s\n",
				req.RemoteAddr, req.URL,
			)
		}
		body, err := ioutil.ReadAll(req.Body)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		if err != nil {
			log.Printf("Error reading body: %v", err)
			return
		}

		if txn.AppendRequestBody(body) != nil {
			log.Println(err.Error())
		}
		if txn.ProcessRequestBody() != nil {
			log.Println(err.Error())
		}
	}

	if txn.ShouldIntervene() {
		log.Printf("Should intervene in request from %s for %s\n",
			req.RemoteAddr, req.URL,
		)
	}
}

func (m *module) loadRules() error {
	ruleset := m.modsec.NewRuleSet()

	m.grpc = pb.NewModuleModsecClient(m.worker.GetGrpcClientConn())
	rulesets, err := m.grpc.GetRules(context.Background(), &empty.Empty{})

	for _, rules := range rulesets.RuleSets {
		if err := ruleset.AddRules(rules.Rules); err != nil {
			return fmt.Errorf("Could not load file '%s': %s", rules.Filename, err.Error())
		}
	}

	m.ruleset = ruleset
	return err
}
