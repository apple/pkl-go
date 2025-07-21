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
	"sync/atomic"
	"unsafe"
)

// MessageHandler is the Go equivalent of PklMessageResponseHandler
// The userData parameter will be the unsafe.Pointer passed to pkl_init
type MessageHandler func(message []byte, userData unsafe.Pointer)

// Global handler (static after initialization)
var globalHandler MessageHandler

//export go_pkl_message_handler_bridge
func go_pkl_message_handler_bridge(length C.int, message *C.char, userData unsafe.Pointer) {
	if globalHandler == nil {
		// Handler not set, this shouldn't happen
		return
	}

	// Convert C data to Go data
	messageBytes := C.GoBytes(unsafe.Pointer(message), length)

	// Call the Go handler with the userData passed from C
	globalHandler(messageBytes, userData)
}

type PklClient struct {
	closed atomic.Bool
	mu     sync.Mutex
}

// New initializes the Pkl executor with a Go callback
func New(handler MessageHandler, userData interface{}) (*PklClient, error) {
	// Convert userData to unsafe.Pointer for C
	var userDataPtr unsafe.Pointer
	if userData != nil {
		userDataPtr = unsafe.Pointer(&userData)
	}

	// Call the C function with our bridge handler
	result := C.pkl_init(C.get_bridge_handler(), userDataPtr)

	if result == -1 {
		// Clean up on failure
		globalHandler = nil
		return nil, errors.New("pkl_init failed")
	}

	// Store the handler globally
	globalHandler = handler

	return &PklClient{}, nil
}

// SendMessage sends a message to Pkl
func (c *PklClient) SendMessage(message []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed.Load() {
		return fmt.Errorf("pkl client is closed")
	}

	if len(message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}

	// Convert Go slice to C data
	cMessage := C.CBytes(message)
	defer C.free(cMessage)

	result := C.pkl_send_message(C.int(len(message)), (*C.char)(cMessage))

	if result == -1 {
		return fmt.Errorf("pkl_send_message failed")
	}

	return nil
}

func (c *PklClient) Version() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed.Load() {
		return "", fmt.Errorf("pkl client is closed")
	}

	version := C.GoString(C.pkl_version())

	return strings.Clone(version), nil
}

// Close cleans up resources
func (c *PklClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed.Load() {
		return nil
	}

	c.closed.Store(true)

	result := C.pkl_close()

	if result == -1 {
		return fmt.Errorf("pkl_close failed")
	}

	return nil
}
