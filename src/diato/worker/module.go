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
	"log"
	"net/http"
	"sync"
)

type moduleInitializer func(*Worker) Module

type Module interface {
	Enabled() bool
	Name() string
	ProcessRequest(*http.Request)
}

type moduleRegistry struct {
	modules []Module
}

var moduleInitializers []func(*Worker) ([]Module, error)

func RegisterModule(initializer func(*Worker) ([]Module, error)) {
	if moduleInitializers == nil {
		moduleInitializers = make([]func(*Worker) ([]Module, error), 0)
	}

	moduleInitializers = append(moduleInitializers, initializer)
}

func (w *Worker) initModules(modules []func(*Worker) ([]Module, error)) error {
	registry := &moduleRegistry{
		modules: make([]Module, 0),
	}
	for _, m := range modules {
		initializedModules, err := m(w)
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

	w.modules = registry
	return nil
}

func (r *moduleRegistry) ProcessRequest(req *http.Request) {
	callbacks := make([]func(), 0)
	for _, m := range r.modules {
		callbacks = append(callbacks, func() { m.ProcessRequest(req) })
	}
	r.parallelCallback(callbacks)
}

func (r *moduleRegistry) parallelCallback(callbacks []func()) chan interface{} {
	out := make(chan interface{}, 8)

	wg := &sync.WaitGroup{}
	for _, callback := range callbacks {
		wg.Add(1)
		go func() {
			callback()
			out <- struct{}{}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}