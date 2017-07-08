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
	"log"
	"net"
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"
	"time"

	"diato/util/stop"
)

func (s *Server) startWorkers(workerCount uint) error {
	httpFd, err := s.getNewHttpSocket()
	if err != nil {
		return err
	}

	throttle := time.Tick(1 * time.Second)
	for i := 1; i <= int(workerCount); i++ {
		if err := s.startWorker(i, httpFd, throttle); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) startWorker(id int, httpFd *os.File, throttle <-chan time.Time) error {
	chrootFd, err := s.getChrootFd()
	if err != nil {
		return err
	}

	cmd := exec.Command(os.Args[0], "internal-worker", "start")
	cmd.ExtraFiles = []*os.File{chrootFd, httpFd}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:    true,
		Pdeathsig: syscall.SIGTERM,
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	atomic.AddInt32(&s.curWorkerCount, 1)

	stopper := stop.NewStopper(func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Process.Wait()
	})

	go func() {
		cmd.Process.Wait()
		remainingWorkerCount := atomic.AddInt32(&s.curWorkerCount, -1)
		if stopper.IsStopping() {
			return
		}

		log.Printf("Worker %d died", id)
		if remainingWorkerCount < 1 {
			panic("No workers remaining, that can't be good. Exiting...")
		}

		<-throttle
		log.Printf("Restarting worker %d...", id)
		s.startWorker(id, httpFd, throttle)
	}()

	return nil
}

// Sets up a new http socket. This socket is used to carry
// plain-text http messages to the worker for further processing
// Messages are supported by the proxy protocol (currently version
// 1) to determine source ip. SSL is stripped in the server daemon.
//
// It is desirable to use version 2 fo the proxy protocol some time
// because it allows to also convey things like SSL usage.
//
// We spawn the socket in the server, the FD is handed over to the
// worker which is responsible for listening and accepting
// connections on this socket. This ensures the worker can run
// in a permission-less environment and does not require any IO.
func (s *Server) getNewHttpSocket() (*os.File, error) {
	listener, err := net.Listen("unix", s.httpSocketPath)
	if err != nil {
		return nil, err
	}

	fd, err := listener.(*net.UnixListener).File()
	if err != nil {
		return nil, err
	}

	stop.NewStopper(func() {
		os.Remove(s.httpSocketPath)
	})
	return fd, nil
}

// Chrooting of the worker is done in plain C where ARGV/ARGC is not
// available. Nor are we going to reimplement the config parser in C,
// as such, we already open and validate the directory that the worker
// should chroot into.
func (s *Server) getChrootFd() (*os.File, error) {
	file, err := os.Open(s.chrootPath)
	if err != nil {
		return nil, err
	}

	fileinfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if !fileinfo.IsDir() {
		return nil, errors.New("Chroot must be a directory, but it does not appear to be")
	}

	return file, nil
}
