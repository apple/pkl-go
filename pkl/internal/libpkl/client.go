//===----------------------------------------------------------------------===//
// Copyright © 2026 Apple Inc. and the Pkl project authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//===----------------------------------------------------------------------===//

//go:build libpkl

package libpkl

/*
#cgo !libpkl_static pkg-config: libpkl
#cgo libpkl_static pkg-config: libpkl-static
#include <stdlib.h>
#include <pkl.h>

// Bridge function to handle Go callbacks from C
// This function will be called by the C library and will forward to Go
void go_pkl_message_handler_bridge(unsigned int length, char *message, void *userData);

// Static C function that acts as the bridge to Go
static void c_pkl_message_handler_bridge(unsigned int length, char *message, void *userData) {
   go_pkl_message_handler_bridge(length, message, userData);
}

// Helper function to get the bridge function pointer
static pkl_message_response_handler get_bridge_handler() {
   return c_pkl_message_handler_bridge;
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	handlerMap sync.Map
	counter    atomic.Int64
)

// MessageHandler is the Go equivalent of PklMessageResponseHandler
type MessageHandler func(message []byte)

//export go_pkl_message_handler_bridge
//goland:noinspection GoSnakeCaseUsage
func go_pkl_message_handler_bridge(length C.uint, message *C.char, userData unsafe.Pointer) {
	id := *(*int64)(userData)
	handler, exists := handlerMap.Load(id)
	if !exists {
		return
	}

	// Convert C data to Go data
	messageBytes := C.GoBytes(unsafe.Pointer(message), C.int(length))

	// Call the Go handler with the original userData provided by the user
	handler.(MessageHandler)(messageBytes)
}

type Id struct {
	id int64
}
type PklClient struct {
	handler MessageHandler
	pinner  *runtime.Pinner
	pexec   *C.pkl_exec_t
	id      int64
	closed  bool
	mu      sync.Mutex
	jobs    chan *job
	stop    chan struct{}
}

type job struct {
	fn   func()
	done chan struct{}
}

// New initializes the Pkl executor with a Go callback
func New(handler MessageHandler) (*PklClient, error) {
	pinner := &runtime.Pinner{}
	id := counter.Add(1)
	idPtr := &id
	// prevent Go from GC'ing this value
	pinner.Pin(idPtr)
	client := &PklClient{
		handler: handler,
		pinner:  pinner,
		id:      id,
		jobs:    make(chan *job),
		stop:    make(chan struct{}),
	}

	go client.run()
	var err error
	client.do(func() {
		var cerr C.pkl_error_t
		var pexec *C.pkl_exec_t

		// Call the C function with our bridge handler
		if ret := C.pkl_init(C.get_bridge_handler(), unsafe.Pointer(idPtr), &pexec, &cerr); ret != 0 {
			err = fmt.Errorf("pkl_init failed: %s", pklErrorMessage(&cerr))
		}
		client.pexec = pexec
	})
	if err != nil {
		return nil, err
	}
	handlerMap.Store(client.id, client.handler)

	return client, nil
}

func (c *PklClient) run() {
	// Don't need to unlock; once goroutine terminates, the Go runtime destroys the underlying OS thread.
	// This is defensive; prevents any thread state mutations on the C/Pkl side from being side-effecting.
	runtime.LockOSThread()
	for {
		select {
		case <-c.stop:
			return
		case j := <-c.jobs:
			j.fn()
			j.done <- struct{}{}
		}
	}
}

func (c *PklClient) do(f func()) {
	j := &job{
		fn:   f,
		done: make(chan struct{}),
	}
	c.jobs <- j
	<-j.done
}

// SendMessage sends a message to Pkl
func (c *PklClient) SendMessage(message []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("pkl client is closed")
	}

	if len(message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}

	// Convert Go slice to C data
	cMessage := C.CBytes(message)
	defer C.free(cMessage)
	var err error
	c.do(func() {
		var cerr C.pkl_error_t
		result := C.pkl_send_message(c.pexec, C.uint(len(message)), (*C.char)(cMessage), &cerr)
		if result != 0 {
			err = fmt.Errorf("pkl_send_message failed: %s", pklErrorMessage(&cerr))
		}
	})
	return err
}

// Close cleans up resources
func (c *PklClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	var err error
	c.do(func() {
		var cerr C.pkl_error_t
		if ret := C.pkl_close(c.pexec, &cerr); ret != 0 {
			err = fmt.Errorf("pkl_close failed: %s", pklErrorMessage(&cerr))
		}
	})
	if err != nil {
		return err
	}
	c.stop <- struct{}{}
	c.closed = true

	handlerMap.Delete(c.id)
	c.pinner.Unpin()

	return nil
}

// pklErrorMessage returns the message contained in a pkl_error_t, or a
// generic fallback if the C library didn't populate one.
func pklErrorMessage(err *C.pkl_error_t) string {
	if err.message == nil {
		return "unknown error"
	}
	return C.GoString(err.message)
}

func Version() string {
	return C.GoString(C.pkl_version())
}
