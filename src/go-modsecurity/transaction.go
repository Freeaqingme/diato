// Diato - Reverse Proxying for Hipsters
//
// Copyright 2016-2017 Dolf Schimmel
// Copyright (c) 2015 Trustwave Holdings, Inc. (http://www.trustwave.com/)
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
*/
import "C"

import (
	"fmt"
	"net"
	"strconv"
	"errors"
	"unsafe"
)

// Represents the inspection on an entire request.
//
// An instance of the transaction struct represents
// an entire request, on its different phases.
type transaction struct {
	ruleset *RuleSet

	//msc_txn *C.struct_transaction
	msc_txn *C.struct_Transaction_t
}

// Create a new transaction for a given configuration and ModSecurity core.
//
// The transaction is the unit that will be used the inspect every request. It holds
// all the information for a given request.
//
// Remember to cleanup the transaction when the transaction is complete. TODO: Describe / implement how
func (r *RuleSet) NewTransaction(remoteAddr, localAddr string) (*transaction, error) {
	remoteIp, remotePort, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("Could not parse remote address: %s", err)
	}

	remotePortInt, err := strconv.Atoi(remotePort)
	if err != nil {
		return nil, fmt.Errorf("Could not convert remote port '%s' to int: %s", remotePort, err.Error())
	}

	localIp, localPort, err := net.SplitHostPort(localAddr)
	if err != nil {
		return nil, fmt.Errorf("Could not parse remote address: %s", err)
	}

	localPortInt, err := strconv.Atoi(localPort)
	if err != nil {
		return nil, fmt.Errorf("Could not convert local port '%s' to int: %s", localPort, err.Error())
	}

	msc_txn := C.msc_new_transaction(r.modsec.modsec, r.msc_rules, nil)
	if msc_txn == nil {
		return nil, fmt.Errorf("Could not initialize transaction")
	}

	cRemoteIp := C.CString(remoteIp) // msc will free() these for us
	cLocalIp := C.CString(localIp)

	// TODO: Check response? @retval 1 Operation was successful.
	if C.msc_process_connection(msc_txn, cRemoteIp, C.int(remotePortInt), cLocalIp, C.int(localPortInt)) != 1 {
		return nil, errors.New("could not process connection")
	}

	return &transaction{
		ruleset: r,
		msc_txn: msc_txn,
	}, nil
}

// Perform the analysis on the URI and all the query string variables.
//
// There is no direct connection between this function and any phase of
// the SecLanguage's phases. It is something that may occur between the
// SecLanguage phase 1 and 2.
func (txn *transaction) ProcessUri(uri, method, httpVersion string) bool {
	cUri := C.CString(uri)
	cMethod := C.CString(method)
	cHttpVersion := C.CString(httpVersion)
	defer C.free(unsafe.Pointer(cUri))
	defer C.free(unsafe.Pointer(cMethod))
	defer C.free(unsafe.Pointer(cHttpVersion))

	// TODO: Check response?
	fmt.Println(uri)
	return 1 ==  C.msc_process_uri(txn.msc_txn, cUri, cMethod, cHttpVersion)
}

// With this function it is possible to feed ModSecurity with a request header.
func (txn *transaction) AddRequestHeader(key, value []byte) bool {
	cKey := (*C.uchar)(unsafe.Pointer(&key[0]))
	cValue := (*C.uchar)(unsafe.Pointer(&value[0]))
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cValue))

	return 1 == C.msc_add_request_header(txn.msc_txn, cKey, cValue)
}

// This function perform the analysis on the request headers, notice however
// that the headers should be added prior to the execution of this function.
//
// Remember to check for a possible intervention.
func (txn *transaction) ProcessRequestHeaders() bool {
	return C.msc_process_request_headers(txn.msc_txn) == 1
}

// Adds request body to be inspected.
//
// With this function it is possible to feed ModSecurity with data for
// inspection regarding the request body.
func (txn *transaction) AppendRequestBody(bodyBuf []byte) bool {
	return 1 == C.msc_append_request_body(txn.msc_txn, (*C.uchar)(unsafe.Pointer(C.CBytes(bodyBuf))), (C.size_t)(len(bodyBuf)) )
}

// Perform the analysis on the request body (if any)
// This function perform the analysis on the request body. It is optional to
// call that function. If this API consumer already know that there isn't a
// body for inspect it is recommended to skip this step.
//
// It is necessary to "append" the request body prior to the execution of this function.
//
// Remember to check for a possible intervention.
func (txn *transaction) ProcessRequestBody() bool {
	return C.msc_process_request_body(txn.msc_txn) == 1
}

func (txn *transaction) Intervention() bool {
	return txn.ShouldIntervene()
}

func (txn *transaction) ShouldIntervene() bool {
	intervention := C.struct_ModSecurityIntervention_t{}
	if C.msc_intervention(txn.msc_txn, &intervention) == 0 {
		fmt.Println("No intervention required!")
		return false
	}

	fmt.Println("INTERVENE!")
	return true
}

func (txn *transaction) Cleanup() {
	C.msc_transaction_cleanup(txn.msc_txn)
	txn.msc_txn = nil
}
