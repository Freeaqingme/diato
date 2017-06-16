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
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"sync"
)

type Filemap struct {
	sync.RWMutex

	path       string
	users      map[string]string
	minEntries int
}

func NewFilemap(path string, entriesRequired int) (*Filemap, error) {
	res := &Filemap{
		path:       path,
		minEntries: entriesRequired,
	}

	err := res.update()
	return res, err
}

func (f *Filemap) update() error {
	contents, err := ioutil.ReadFile(f.path)
	if err != nil {
		return fmt.Errorf("Could not read file %s: %s", f.path, err.Error())
	}
	return f.updateWithContents(contents)
}

func (f *Filemap) updateWithContents(contents []byte) error {
	lines := bytes.Split(contents, []byte("\n"))

	newMap := make(map[string]string)
	for i, line := range lines {
		lineParts := bytes.Split(bytes.TrimSpace(line), []byte(" "))
		if len(lineParts[0]) == 0 {
			continue
		}

		user := string(lineParts[:1][0])
		host := string(lineParts[len(lineParts)-1:][0])

		if _, alreadyExists := newMap[user]; alreadyExists {
			return fmt.Errorf("Domain %s was defined more than once on line %d", user, i)
		}
		newMap[user] = host
	}

	if len(newMap) < f.minEntries {
		return fmt.Errorf("New Map only contains %d entries, which is less than the set minimum %d",
			len(newMap), f.minEntries)
	}

	f.Lock()
	defer f.Unlock()
	f.users = newMap

	return nil
}

func (f *Filemap) GetServerForUser(user string) (string, uint32, error) {
	f.RLock()
	entry, exists := f.users[user]
	f.RUnlock()

	if !exists {
		return "", 0, fmt.Errorf("No mapping could be found for user '%s'", user)
	}

	host, portStr, err := net.SplitHostPort(entry)
	if err != nil {
		return host, 0, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return host, 0, fmt.Errorf("Could not parse port from file map: %s", err.Error())
	}

	return host, uint32(port), nil
}

func (f *Filemap) getNoOfUsers() int {
	f.RLock()
	defer f.RUnlock()

	return len(f.users)
}
