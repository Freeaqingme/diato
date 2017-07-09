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

// Inspired by Cockroachdb's stop util.
package stop

import (
	"sync"
)

func register(s *Stopper) {
	trackedStoppers.Lock()
	trackedStoppers.stoppers = append(trackedStoppers.stoppers, s)
	trackedStoppers.Unlock()
}

var trackedStoppers struct {
	sync.Mutex
	stoppers []*Stopper
	stopped  bool
}

type Stopper struct {
	stopper  chan struct{}
	stopping bool
	callback func()
}

func Stop() {
	trackedStoppers.Lock()
	trackedStoppers.stopped = true

	for _, stopper := range trackedStoppers.stoppers {
		stopper.Stop()
	}
	trackedStoppers.Unlock()
}

func NewStopper(callback func()) *Stopper {
	s := &Stopper{
		make(chan struct{}, 0),
		false,
		callback,
	}
	register(s)

	trackedStoppers.Lock()
	if trackedStoppers.stopped {
		trackedStoppers.Unlock()
		s.Stop()
	} else {
		trackedStoppers.Unlock()
	}
	return s
}

func (s *Stopper) ShouldStop() <-chan struct{} {
	if s == nil {
		return nil
	}
	return s.stopper
}

func (s *Stopper) Stop() {
	if s.stopping {
		return
	}
	s.stopping = true
	close(s.stopper)
	if s.callback != nil {
		s.callback()
	}
}

func (s *Stopper) IsStopping() bool {
	return s.stopping
}
