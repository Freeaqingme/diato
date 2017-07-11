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

	config "diato/config"
)

type moduleInitializer func(*Worker) Module

type Module interface {
	Enabled() bool
	Name() string
	ProcessRequest(*http.Request)
	PostModifyResponse(w *http.Response)
}

type moduleRegistry struct {
	modules []Module
}

var moduleInitializers []func(*Worker, *config.Config) ([]Module, error)

func RegisterModule(initializer func(*Worker, *config.Config) ([]Module, error)) {
	if moduleInitializers == nil {
		moduleInitializers = make([]func(*Worker, *config.Config) ([]Module, error), 0)
	}

	moduleInitializers = append(moduleInitializers, initializer)
}

func (w *Worker) initModules(modules []func(*Worker, *config.Config) ([]Module, error), config *config.Config) error {
	registry := &moduleRegistry{
		modules: make([]Module, 0),
	}
	for _, m := range modules {
		initializedModules, err := m(w, config)
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
		callback := m.ProcessRequest
		callbacks = append(callbacks, func() { (callback)(req) })
	}
	r.parallelCallback(callbacks)
}

func (r *moduleRegistry) PostModifyResponse(resp *http.Response) {
	callbacks := make([]func(), 0)
	for _, m := range r.modules {
		callback := m.PostModifyResponse
		callbacks = append(callbacks, func() { (callback)(resp) })
	}
	r.parallelCallback(callbacks)
}

func (r *moduleRegistry) parallelCallback(callbacks []func()) chan interface{} {
	out := make(chan interface{}, 8)

	wg := &sync.WaitGroup{}
	for _, callback := range callbacks {
		wg.Add(1)
		go func(callback func()) {
			callback()
			out <- struct{}{}
			wg.Done()
		}(callback)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

type ModuleBase struct {
}

func (*ModuleBase) ProcessRequest(*http.Request)      {}
func (*ModuleBase) PostModifyResponse(*http.Response) {}
