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

	"diato/module/elasticsearch/worker/mapping/v1"
)

type mapping interface {
	PersistRequest(*http.Request, *http.Response)
	EnsureTemplate() error
}

func (m *module) getActiveMapping() mapping {
	if m.activeMapping == nil {
		m.activeMapping = v1.NewMapping(m.client)
	}
	return m.activeMapping
}

func (m *module) persistRequest(req *http.Request, resp *http.Response) {
	m.getActiveMapping().PersistRequest(req, resp)
}

func (m *module) ensureTemplate() error {
	return m.getActiveMapping().EnsureTemplate()
}
