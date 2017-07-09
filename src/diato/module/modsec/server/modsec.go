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
	"diato/server"
	"fmt"
	"io/ioutil"

	"diato/module/modsec/pb"
	"github.com/mattn/go-zglob"
)

const name = "modsec"

func init() {
	server.RegisterModule(newModule)
}

type module struct {
	enabled bool

	rules *pb.RuleSets
}

func newModule(w *server.Server, config *server.Config) ([]server.Module, error) {
	module := &module{
		enabled: config.Modsec.Enabled,
		rules: &pb.RuleSets{
			RuleSets: make([]*pb.RuleSet, 0),
		},
	}

	if !module.Enabled() {
		return []server.Module{module}, nil
	}

	if err := module.loadRulePaths(config.Modsec.RulesFile); err != nil {
		return []server.Module{}, err
	}

	return []server.Module{module}, nil
}

func (m *module) Enabled() bool {
	return m.enabled
}

func (m *module) Name() string {
	return name
}

func (m *module) loadRulePaths(paths []string) error {
	for _, path := range paths {
		if err, errPath := m.loadRulePath(path); err != nil {
			return fmt.Errorf("Could not open path '%s': %s", errPath, err.Error())
		}
	}

	return nil
}

func (m *module) loadRulePath(globPath string) (error, string) {
	paths, err := zglob.Glob(globPath)
	if err != nil {
		return err, globPath
	}

	origCount := len(m.rules.RuleSets)
	for _, path := range paths {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return err, path
		}

		m.rules.RuleSets = append(m.rules.RuleSets, &pb.RuleSet{
			Rules:    string(contents),
			Filename: path,
		})
	}

	if origCount == len(m.rules.RuleSets) {
		return fmt.Errorf("No rule files found"), globPath
	}

	return nil, ""
}
