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

	"diato/userbackend"
	"diato/userbackend/filemap"
)

type Server struct {
	userBackend userbackend.Userbackend

	httpSocketPath string
	chrootPath     string
	tlsCertDir     string

	workerLimit uint

	tlsCertStore   *tlsCertStore
	curWorkerCount int32
}

func Start(config *Config) error {
	s := &Server{
		httpSocketPath: config.General.HttpSocketPath,
		chrootPath:     config.General.Chroot,
		tlsCertDir:     config.General.TlsCertDir,
	}

	if !config.FilemapUserbackend.Enabled {
		return errors.New("No user backends were enabled")
	}

	userbackendConfig := config.FilemapUserbackend
	var err error
	s.userBackend, err = Filemap.NewFilemap(userbackendConfig.Path, userbackendConfig.MinEntries)
	if err != nil {
		return fmt.Errorf("Could ont initialize filemap userbackend: %s", err.Error())
	}

	if err := s.startRpc(); err != nil {
		return err
	}

	if err := s.startWorkers(config.General.WorkerCount); err != nil {
		return err
	}

	for _, listen := range config.Listen {
		if err := s.Listen(listen.Bind, listen.TlsEnable); err != nil {
			return err
		}
	}

	return nil
}
