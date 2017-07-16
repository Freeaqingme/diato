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
	"errors"
	"fmt"

	"diato/config"
	"diato/userbackend"
	"diato/userbackend/filemap"

	"gopkg.in/gcfg.v1"
	"io/ioutil"
)

type Server struct {
	userBackend userbackend.Userbackend

	httpSocketPath  string
	httpsSocketPath string
	chrootPath      string
	tlsCertDir      string

	workerLimit uint

	tlsCertStore   *tlsCertStore
	curWorkerCount int32
	modules        *moduleRegistry

	// configFileContents contains the contents of the
	// config file as it was read on start-up. This is
	// not used in the server other than to pass it on
	// to the worker when requested.
	configFileContents []byte

	httpBind []httpBind
}

func Start(configPath string) error {
	s, config, err := newServer(configPath)
	if err != nil {
		return err
	}

	if !config.FilemapUserbackend.Enabled {
		return errors.New("No user backends were enabled")
	}

	userbackendConfig := config.FilemapUserbackend
	s.userBackend, err = Filemap.NewFilemap(userbackendConfig.Path, userbackendConfig.MinEntries)
	if err != nil {
		return fmt.Errorf("Could ont initialize filemap userbackend: %s", err.Error())
	}

	if err := s.initModules(moduleInitializers, config); err != nil {
		return err
	}

	if err := s.startRpc(); err != nil {
		return err
	}

	if err := s.startWorkers(config.General.WorkerCount); err != nil {
		return err
	}

	for name, l := range config.Listen {
		bind := httpBind{
			name:       name,
			listen:     l.Bind,
			proxyProto: l.ProxyProtocol,
			hasSsl:     l.TlsEnable,
		}
		s.httpBind = append(s.httpBind, bind)
		if err := s.Listen(&bind); err != nil {
			return err
		}
	}

	return nil
}

func newServer(configPath string) (*Server, *config.Config, error) {
	configFileContents, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not open config file: %s", err.Error())
	}

	config := config.NewConfig()
	err = gcfg.ReadFileInto(config, configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not parse configuration: %s", err.Error())
	}

	if err = config.Validate(); err != nil {
		return nil, nil, fmt.Errorf("Could not parse configuration: %s", err.Error())
	}

	s := &Server{
		httpSocketPath:     config.General.HttpSocketPath,
		httpsSocketPath:    config.General.HttpsSocketPath,
		chrootPath:         config.General.Chroot,
		tlsCertDir:         config.General.TlsCertDir,
		configFileContents: configFileContents,
	}
	return s, config, nil
}
