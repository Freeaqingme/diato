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
package elasticsearch

import (
	"net/http"

	"diato/config"
	"diato/worker"

	"gopkg.in/olivere/elastic.v5"
)

const name = "elasticsearch"

func init() {
	worker.RegisterModule(newModule)
}

type module struct {
	enabled bool

	*worker.ModuleBase
	worker *worker.Worker

	client        *elastic.Client
	activeMapping mapping
}

func newModule(w *worker.Worker, config *config.Config) ([]worker.Module, error) {
	module := &module{
		worker:  w,
		enabled: config.Elasticsearch.Enabled,
	}

	if !config.Elasticsearch.Enabled {
		return []worker.Module{module}, nil
	}

	var err error
	module.client, err = elastic.NewClient(
		elastic.SetSniff(config.Elasticsearch.Sniff),
		elastic.SetURL(config.Elasticsearch.Url...),
	)
	if err != nil {
		return nil, err
	}

	if err := module.ensureTemplate(); err != nil {
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

func (m *module) PostModifyResponse(req *http.Request, resp *http.Response) {
	m.persistRequest(req, resp)
}
