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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"diato/util/stop"

	"github.com/rjeczalik/notify"
)

type tlsCertStore struct {
	*sync.RWMutex

	config *tls.Config

	nameToCert map[string][]*tlsCert
	pathToCert map[string]*tlsCert
}

type tlsCert struct {
	tls.Certificate

	loadedAt time.Time
	names    []string
	path     string
}

func (s *Server) tlsListen(ln net.Listener) (net.Listener, error) {
	tlsConf, err := s.tlsGetConfig()
	if err != nil {
		return nil, err
	}
	ln = tls.NewListener(ln, tlsConf)

	return ln, nil
}

func (s *Server) tlsGetConfig() (*tls.Config, error) {
	if s.tlsCertStore == nil {
		certStore := &tlsCertStore{
			RWMutex:    &sync.RWMutex{},
			config:     &tls.Config{},
			nameToCert: make(map[string][]*tlsCert, 0),
			pathToCert: make(map[string]*tlsCert, 0),
		}
		certStore.config.GetCertificate = certStore.getCertificate

		if err := certStore.watchForUpdates(s.tlsCertDir); err != nil {
			return nil, err
		}
		if err := certStore.loadCertsFromFilesystem(s.tlsCertDir); err != nil {
			return nil, err
		}

		log.Printf("Loaded %d certificates for %d names", certStore.NumberOfCerts(), certStore.NumberOfNames())

		s.tlsCertStore = certStore
	}

	return s.tlsCertStore.config, nil
}

func (s *tlsCertStore) loadCertsFromFilesystem(searchDir string) error {
	// Unfortunately there's no way (yet) to properly clean up stoppers,
	// as such, this one will continue lingering until the end of the
	// application lifespan. Which is a bit of a pita, but if it works...
	stopper := stop.NewStopper(nil)

	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if stopper.IsStopping() {
			return nil // No way to abort walking, but we surely can speed it up by not parsing anything
		}
		if f.IsDir() {
			return nil
		}

		name := f.Name()
		if !hasPemExtension(name) {
			return nil
		}

		if err := s.loadCertFromFilesystem(path); err != nil {
			return errors.New("Could not load certificate from '" + path + "': " + err.Error())
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// getCertificate returns the best certificate for the given ClientHelloInfo.
// We override it from tls.getCertificate() because we want to maintain our
// own cert store so we can do some proper locking when we add more certs
// while the listener is already active - a use case Go's TLS library appears
// not to have been designed for.
func (s *tlsCertStore) getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	s.RLock()
	defer s.RUnlock()

	name := strings.ToLower(clientHello.ServerName)
	for len(name) > 0 && name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}

	if cert, ok := s.nameToCert[name]; ok {
		return &cert[0].Certificate, nil
	}

	labels := strings.Split(name, ".")
	for i := range labels {
		labels[i] = "*"
		candidate := strings.Join(labels, ".")
		if cert, ok := s.nameToCert[candidate]; ok {
			return &cert[0].Certificate, nil
		}
	}

	// If nothing matches, return nothing
	return nil, nil
	//return &s.certs[0], nil
}

func (s *tlsCertStore) NumberOfCerts() int {
	s.RLock()
	defer s.RUnlock()

	return len(s.pathToCert)
}

func (s *tlsCertStore) NumberOfNames() int {
	s.RLock()
	defer s.RUnlock()

	return len(s.nameToCert)
}

func (s *tlsCertStore) watchForUpdates(path string) error {
	c := make(chan notify.EventInfo, 1024)

	// Set up a watchpoint listening for events within a directory tree rooted
	// at current working directory. Dispatch remove events to c.
	if err := notify.Watch(path+"/...", c, notify.Remove, notify.Create, notify.Write, notify.Rename); err != nil {
		return err
	}

	stopper := stop.NewStopper(func() {
		notify.Stop(c)
	})

	go func() {
		for {
			select {
			case event := <-c:
				s.handlePemFileEvent(event)
			case _ = <-stopper.ShouldStop():
				return
			}
		}
	}()
	return nil
}

func (s *tlsCertStore) handlePemFileEvent(event notify.EventInfo) {
	switch event.Event() {
	case notify.Rename:
		// Will trigger a separate Create if it's in the target dir
		fallthrough
	case notify.Remove:
		s.removeCertByPath(event.Path())
		log.Printf("Removed cert '%s', %d certificates for %d names remaining",
			filepath.Base(event.Path()), s.NumberOfCerts(), s.NumberOfNames())
	case notify.Create:
		s.loadCertFromFilesystem(event.Path())
		log.Printf("Loaded cert '%s', now have %d certificates for %d names",
			filepath.Base(event.Path()), s.NumberOfCerts(), s.NumberOfNames())
	case notify.Write:
		s.removeCertByPath(event.Path())
		s.loadCertFromFilesystem(event.Path())
		log.Printf("Updated cert '%s', still have %d certificates for %d names",
			filepath.Base(event.Path()), s.NumberOfCerts(), s.NumberOfNames())
	}
}

func (s *tlsCertStore) loadCertFromFilesystem(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return errors.New("Could not determine absolute path to cert: " + err.Error())
	}

	cert, err := tls.LoadX509KeyPair(path, path)
	if err != nil {
		return err
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}

	decoratedCert := &tlsCert{
		Certificate: cert,
		loadedAt:    time.Now(),
		path:        path,
	}

	s.Lock()
	defer s.Unlock()

	names := make([]string, 0)
	if len(x509Cert.Subject.CommonName) > 0 {
		names = append(names, x509Cert.Subject.CommonName)
		if mapEntry, exists := s.nameToCert[x509Cert.Subject.CommonName]; exists {
			s.nameToCert[x509Cert.Subject.CommonName] = append(mapEntry, decoratedCert)
		} else {
			s.nameToCert[x509Cert.Subject.CommonName] = []*tlsCert{decoratedCert}
		}
	}
	for _, san := range x509Cert.DNSNames {
		names = append(names, san)
		if mapEntry, exists := s.nameToCert[san]; exists {
			s.nameToCert[san] = append(mapEntry, decoratedCert)
		} else {
			s.nameToCert[san] = []*tlsCert{decoratedCert}
		}
	}

	decoratedCert.names = names
	s.pathToCert[path] = decoratedCert
	return nil
}

func (s *tlsCertStore) removeCertByPath(path string) {
	if !hasPemExtension(path) {
		return
	}

	s.Lock()
	cert, exists := s.pathToCert[path]
	if !exists {
		log.Printf("A cert was removed but we didn't have it in the first place: %s", path)
		s.Unlock()
		return
	}

	for _, name := range cert.names {
		namedCerts, ok := s.nameToCert[name]
		if !ok {
			continue
		}

		for i := range namedCerts {
			if namedCerts[i].path != cert.path {
				continue
			}
			namedCerts[i] = namedCerts[len(namedCerts)-1]
			namedCerts[len(namedCerts)-1] = nil
			s.nameToCert[name] = namedCerts[:len(namedCerts)-1]
			break
		}
		if len(s.nameToCert[name]) == 0 {
			delete(s.nameToCert, name)
		}
	}

	delete(s.pathToCert, path)
	s.Unlock()
}

func hasPemExtension(name string) bool {
	if len(name) <= 5 {
		return false
	}

	if name[len(name)-4:] != ".pem" {
		return false
	}

	return true
}
