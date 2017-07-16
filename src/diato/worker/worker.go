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

/*
#cgo CFLAGS: -Wall
#cgo LDFLAGS: -lcap -lseccomp

extern void secureEnvironment();
void __attribute__((constructor)) init(void) {
	secureEnvironment();
}
*/
import "C"

import (
	"fmt"
	"os"

	"diato/config"
	"diato/pb"

	"google.golang.org/grpc"
	"gopkg.in/gcfg.v1"
)

type Worker struct {
	userBackend diato.UserBackendClient

	modules        *moduleRegistry
	grpcClientConn *grpc.ClientConn
}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) Start() error {
	if os.Getuid() == 0 {
		panic("The worker refuses to run as root profusely. " +
			"Don't invoke it manually, just use 'daemon start'")
	}

	var err error
	if w.grpcClientConn, err = w.rpcInit(); err != nil {
		return err
	}

	config, err := w.getConfig()
	if err != nil {
		return err
	}

	if err := w.initModules(moduleInitializers, config); err != nil {
		return err
	}

	httpListener, err := w.httpGetListener(false)
	if err != nil {
		return err
	}
	go w.httpListen(httpListener, false)

	httpsListener, err := w.httpGetListener(true)
	if err != nil {
		return err
	}
	go w.httpListen(httpsListener, true)

	return nil
}

func (w *Worker) getConfig() (*config.Config, error) {
	configContents, err := w.getConfigContents()
	if err != nil {
		return nil, err
	}

	config := config.NewConfig()
	err = gcfg.ReadStringInto(config, string(configContents))
	if err != nil {
		return nil, fmt.Errorf("Could not parse configuration: %s", err.Error())
	}

	if err = config.Validate(); err != nil {
		return nil, fmt.Errorf("Could not parse configuration: %s", err.Error())
	}

	return config, nil
}
