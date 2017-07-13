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
package config

import (
	"errors"

	"diato/userbackend/filemap"

	elasticsearch "diato/module/elasticsearch/worker/config"
	modsec "diato/module/modsec/server/config"
)

type Config struct {
	General            GeneralConfig  `gcfg:"diato"`
	FilemapUserbackend Filemap.Config `gcfg:"filemap-userbackend"`
	Listen             map[string]*struct {
		Bind      string
		TlsEnable bool `gcfg:"tls-enable"`
	}

	Elasticsearch elasticsearch.Config `gcfg:"elasticsearch"`
	Modsec        modsec.Config        `gcfg:"modsecurity"`
}

type GeneralConfig struct {
	HttpSocketPath string `gcfg:"http-socket-path"`
	Chroot         string
	TlsCertDir     string `gcfg:"tls-cert-dir"`
	WorkerCount    uint   `gcfg:"worker-count"`
}

func NewConfig() *Config {
	return &Config{
		General: GeneralConfig{
			HttpSocketPath: "/var/run/diato/http.socket",
			Chroot:         "/var/run/diato/chroot",
		},
	}
}

func (c *Config) Validate() error {
	if len(c.Listen) == 0 {
		return errors.New("No listen sections defined, expected at least one")
	}

	return nil
}
