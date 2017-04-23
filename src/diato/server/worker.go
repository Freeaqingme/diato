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
	"fmt"

	"math/rand"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	"diato/util/stop"
)

func (s *Server) startWorker() error {
	fd, err := s.getNewWorkerSocket()
	if err != nil {
		return err
	}

	cmd := exec.Command(os.Args[0], "internal-worker", "start")
	cmd.ExtraFiles = []*os.File{fd}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: 65534,
			Gid: 65534,
		},
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	stopper := stop.NewStopper(func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Process.Wait()
	})

	go func() {
		cmd.Process.Wait()
		if !stopper.IsStopping() {
			panic("Worker died!")
		}
	}()
	return nil
}

func (s *Server) getNewWorkerSocket() (*os.File, error) {
	rand.Seed(time.Now().UnixNano())
	s.workerSocketPath = fmt.Sprintf("/tmp/diato-worker.%d", rand.Int())
	listener, err := net.Listen("unix", s.workerSocketPath)
	if err != nil {
		return nil, err
	}

	fd, err := listener.(*net.UnixListener).File()
	if err != nil {
		return nil, err
	}

	stop.NewStopper(func() {
		os.Remove(s.workerSocketPath)
	})
	return fd, nil
}
