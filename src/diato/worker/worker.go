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
#cgo LDFLAGS: -L/usr/include/sys/capability.h

extern void secureEnvironment();
void __attribute__((constructor)) init(void) {
	secureEnvironment();
}
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"diato/pb"
	seccomp "github.com/seccomp/libseccomp-golang"
)

type Worker struct {
	userBackend diato.UserBackendClient
}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) Start() error {
	if os.Getuid() == 0 {
		return errors.New("The worker refuses to run as root profusely. " +
			"Don't invoke it manually, just use 'daemon start'")
	}

	httpListener, err := w.httpGetListener()
	if err != nil {
		return err
	}
	go w.httpListen(httpListener)

	if err := w.rpcInit(); err != nil {
		return err
	}

	// TODO: Check uid as to ensure the dropping of privileges actually ran

	return nil

	// See: https://github.com/seccomp/libseccomp-golang/issues/23#issuecomment-296441184
	fmt.Println("Setting seccomp")
	if err := w.seccomp(); err != nil {
		fmt.Println("Error while loading filter")
		fmt.Println(err)
		return err
	}
	fmt.Println("Seccomp set")
	time.Sleep(5 * time.Second)
	return nil
}

func (s *Worker) seccomp() error {
	filter, err := seccomp.NewFilter(seccomp.ActKill)
	filter.AddRule(seccomp.ScmpSyscall(syscall.SYS_MADVISE), seccomp.ActKill)
	if err != nil {
		return err
	}
	fmt.Println("about to load filter")
	return filter.Load()
}
