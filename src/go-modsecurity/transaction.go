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


int msc_add_request_header_bridge(Transaction *transaction, char *key, char *value) {
	unsigned char * ukey;
	unsigned char * uvalue;
	ukey = (unsigned char *) key;
	uvalue = (unsigned char *) value;

	int ret;
	ret = msc_add_request_header(transaction, ukey, uvalue);
    return ret;
}

*/
import "C"

import (
	"fmt"
	"net"
	"strconv"
	"net/http"
	"errors"
	"unsafe"
)

type transaction struct {
	ruleset *RuleSet

	//msc_txn *C.struct_transaction
	msc_txn *C.struct_Transaction_t
}

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

func (txn *transaction) ProcessUri(uri, method, httpVersion string) bool {
	// TODO: Check response?
	fmt.Println(uri)
	return 1 ==  C.msc_process_uri(txn.msc_txn,
		C.CString(uri), C.CString(method), C.CString(httpVersion))
}

func (txn *transaction) AddRequestHeader(key, value string) bool {
	return C.msc_add_request_header_bridge(txn.msc_txn, C.CString(key), C.CString(value) ) == 1

}

func (txn *transaction) ProcessRequestHeaders(hdrs *http.Header) bool {
	if hdrs == nil {
		goto process
	}

	for key,values := range *hdrs {
		for _, value := range values {
			txn.AddRequestHeader(key, value)
		}
	}

process:
	return C.msc_process_request_headers(txn.msc_txn) == 1
}

func (txn *transaction) AppendRequestBody(bodyBuf []byte) bool {
	return C.msc_append_request_body(txn.msc_txn, (*C.uchar)(unsafe.Pointer(C.CBytes(bodyBuf))), (C.size_t)(len(bodyBuf)) )== 1
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

/*
msc_process_request_body(transaction);
msc_add_response_header(transaction, "Content-type", "text/html");
msc_process_response_headers(transaction, 200, "HTTP 1.0");
msc_process_response_body(transaction);
msc_process_logging(transaction);
msc_transaction_cleanup(transaction);

*/
