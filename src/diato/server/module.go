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
	"log"

	"diato/config"

	"google.golang.org/grpc"
)

type moduleInitializer func(*Server, *config.Config) Module

type Module interface {
	Enabled() bool
	Name() string
	RegisterRpcEndpoints(*grpc.Server)
}

type moduleRegistry struct {
	modules []Module
}

var moduleInitializers []func(*Server, *config.Config) ([]Module, error)

func RegisterModule(initializer func(*Server, *config.Config) ([]Module, error)) {
	if moduleInitializers == nil {
		moduleInitializers = make([]func(*Server, *config.Config) ([]Module, error), 0)
	}

	moduleInitializers = append(moduleInitializers, initializer)
}

func (s *Server) initModules(modules []func(*Server, *config.Config) ([]Module, error), config *config.Config) error {
	registry := &moduleRegistry{
		modules: make([]Module, 0),
	}
	for _, m := range modules {
		initializedModules, err := m(s, config)
		if err != nil {
			return err
		}
		for _, initializedModule := range initializedModules {
			if !initializedModule.Enabled() {
				log.Printf("Skipping module '%s' because it was not enabled", initializedModule.Name())
				continue
			}
			log.Printf("Loaded module '%s'", initializedModule.Name())
			registry.modules = append(registry.modules, initializedModule)
		}
	}

	s.modules = registry
	return nil
}
