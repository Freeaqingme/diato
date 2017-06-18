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
package Filemap

import (
	"errors"
	"log"

	"diato/util/stop"

	"github.com/rjeczalik/notify"
)

func (f *Filemap) watchForUpdates() error {
	f.stopWatcherIfRunning()

	f.watcher = make(chan notify.EventInfo, 1)
	if err := notify.Watch(f.path, f.watcher, notify.Write, notify.Remove); err != nil {
		return errors.New("Could not instantiate new watcher on user map:" + err.Error())
	}

	stopper := stop.NewStopper(f.stopWatcherIfRunning)
	go func() {
		for {
			select {
			case event := <-f.watcher:
				if f.handleFilemapFileUpdate(event) {
					return
				}
			case _ = <-stopper.ShouldStop():
				return
			}
		}
	}()

	return nil
}

func (f *Filemap) handleFilemapFileUpdate(event notify.EventInfo) (abort bool) {
	switch event.Event() {
	case notify.Remove:
		f.watchForUpdates() // Set new watcher
		if err := f.update(); err != nil {
			log.Print("Error: Could not update usermap file", err.Error())
		}
		return true
	default:
		if err := f.update(); err != nil {
			log.Print("Error: Could not update usermap file", err.Error())
		}
		return false
	}
}

func (f *Filemap) stopWatcherIfRunning() {
	if f.watcher == nil {
		return
	}
	notify.Stop(f.watcher)
	f.watcher = nil
}
