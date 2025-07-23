//===----------------------------------------------------------------------===//
// Copyright © 2025 Apple Inc. and the Pkl project authors. All rights reserved.
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
#cgo LDFLAGS: -lpkl -lpkl_internal
#include <stdlib.h>
#include <pkl.h>

// Bridge function to handle Go callbacks from C
// This function will be called by the C library and will forward to Go
void go_pkl_message_handler_bridge(int length, char *message, void *userData);

// Static C function that acts as the bridge to Go
static void c_pkl_message_handler_bridge(int length, char *message, void *userData) {
   go_pkl_message_handler_bridge(length, message, userData);
}

// Helper function to get the bridge function pointer
static PklMessageResponseHandler get_bridge_handler() {
   return c_pkl_message_handler_bridge;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"github.com/google/uuid"
)

var handlerMap sync.Map

// MessageHandler is the Go equivalent of PklMessageResponseHandler
// The userData parameter will be the unsafe.Pointer passed to pkl_init
type MessageHandler func(message []byte, userData unsafe.Pointer)

//export go_pkl_message_handler_bridge
func go_pkl_message_handler_bridge(length C.int, message *C.char, userData unsafe.Pointer) {
	handler, exists := handlerMap.Load(userData)
	if !exists {
		return
	}

	// Convert C data to Go data
	messageBytes := C.GoBytes(unsafe.Pointer(message), length)

	// Call the Go handler with the original userData provided by the user
	handler.(MessageHandler)(messageBytes, userData)
}

type PklClient struct {
	handler  MessageHandler
	pexec    *C.pkl_exec_t
	userData interface{}

	id        uuid.UUID
	idPointer unsafe.Pointer

	closed bool
	mu     sync.Mutex
}

// New initializes the Pkl executor with a Go callback
func New(handler MessageHandler) (*PklClient, error) {
	uuid := uuid.New()

	client := &PklClient{
		handler:   handler,
		id:        uuid,
		idPointer: unsafe.Pointer(&uuid),
	}

	// Call the C function with our bridge handler
	pexec := C.pkl_init(C.get_bridge_handler(), client.idPointer)
	if pexec == nil {
		return nil, errors.New("pkl_init failed")
	}

	client.pexec = pexec
	handlerMap.Store(client.idPointer, client.handler)

	return client, nil
}

// export go_pkl_message_handler
func (c *PklClient) messageHandler(length C.int, message *C.char, userData unsafe.Pointer) {
	messageBytes := C.GoBytes(unsafe.Pointer(message), length)
	c.handler(messageBytes, userData)
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

	result := C.pkl_send_message(c.pexec, C.int(len(message)), (*C.char)(cMessage))

	if result == -1 {
		return fmt.Errorf("pkl_send_message failed")
	}

	return nil
}

func (c *PklClient) Version() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return "", fmt.Errorf("pkl client is closed")
	}

	version := C.GoString(C.pkl_version(c.pexec))

	return strings.Clone(version), nil
}

// Close cleans up resources
func (c *PklClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	result := C.pkl_close(c.pexec)
	if result == -1 {
		return fmt.Errorf("pkl_close failed")
	}

	handlerMap.Delete(c.id)

	return nil
}
