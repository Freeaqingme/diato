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

package modsecurity

/*
#cgo CPPFLAGS: -I/usr/local/modsecurity/include
#cgo LDFLAGS: /usr/local/modsecurity/lib/libmodsecurity.so

#include "modsecurity/modsecurity.h"
#include "modsecurity/transaction.h"
#include "modsecurity/rules.h"

*/
import "C"

import (
	"errors"
)

type Modsecurity struct {
	modsec *C.struct_ModSecurity_t
}

func NewModsecurity() (*Modsecurity, error) {
	modsec := C.msc_init()

	if modsec == nil {
		return nil, errors.New("Could not initialize Mod Security")
	}

	return &Modsecurity{
		modsec: modsec,
	}, nil
}
