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
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"diato/worker"

	"gopkg.in/olivere/elastic.v5"
)

type mapping struct {
	client *elastic.Client
}

func NewMapping(client *elastic.Client) *mapping {
	return &mapping{client}
}

func (m *mapping) EnsureTemplate() error {
	_, err := m.client.IndexPutTemplate("diato-httprequest-v1").
		BodyString(template).
		Do(context.TODO())

	return err
}

func (m *mapping) PersistRequest(req *http.Request, resp *http.Response) {
	ctx := req.Context()
	ctxInfo := ctx.Value("diato").(*worker.ContextInfo)

	request := m.getStructuredRequest(req, resp)
	_, err := m.client.Index().
		Index(fmt.Sprintf("diato-httprequest-%s-v1", time.Now().Format("20060102"))).
		Type("httprequest").
		Id(ctxInfo.RequestIdString()).
		BodyJson(request).
		Refresh("true").
		Do(ctx)
	if err != nil {
		log.Printf("Error while inserting into ElasticSearch for request %s: %s",
			ctxInfo.RequestIdString(), err.Error())
	}
}
