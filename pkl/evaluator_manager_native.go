//go:build libpkl

package pkl

import (
	"bytes"
	"fmt"
	"github.com/apple/pkl-go/pkl/internal"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"sync"
	"unsafe"

	"github.com/apple/pkl-go/pkl/internal/msgapi"
	"github.com/apple/pkl-go/pkl/libpkl"
)

var _ evaluatorManagerImpl = (*nativeEvaluator)(nil)

// NewEvaluatorManager creates a new EvaluatorManager using the `libpkl` native bindings.
func NewEvaluatorManager() EvaluatorManager {
	m := &evaluatorManager{
		impl: &nativeEvaluator{
			in:       make(chan msgapi.IncomingMessage),
			out:      make(chan msgapi.OutgoingMessage),
			received: make(chan []byte),
			closed:   make(chan error),
		},
		interrupts:        &sync.Map{},
		evaluators:        &sync.Map{},
		pendingEvaluators: &sync.Map{},
	}
	go m.listen()
	go m.listenForImplClose()
	return m
}

type nativeEvaluator struct {
	client   *libpkl.PklClient
	in       chan msgapi.IncomingMessage
	out      chan msgapi.OutgoingMessage
	received chan []byte
	closed   chan error

	// exited is a flag that indicates evaluator was closed explicitly
	exited  atomicBool
	version *semver
}

func (n *nativeEvaluator) init() error {
	c, err := libpkl.New(n.responseHandler, nil)
	if err != nil {
		return err
	}
	n.client = c

	go n.handleSendMessages()

	return nil
}

func (n *nativeEvaluator) deinit() error {
	n.exited.set(true)

	close(n.closed)
	close(n.in)
	close(n.out)
	close(n.received)

	// TODO: Figure out how to close the native library down cleanly as part of the tests.
	// return n.client.Close()

	return nil
}

func (n *nativeEvaluator) inChan() chan msgapi.IncomingMessage { return n.in }

func (n *nativeEvaluator) outChan() chan msgapi.OutgoingMessage { return n.out }

func (n *nativeEvaluator) closedChan() chan error { return n.closed }

func (n *nativeEvaluator) getVersion() (*semver, error) {
	if n.exited.get() {
		return nil, fmt.Errorf("evaluator is closed")
	}

	version, err := n.client.Version()
	if err != nil {
		return nil, err
	}

	parsed, err := parseSemver(version)
	if err != nil {
		return nil, err
	}
	n.version = parsed
	return n.version, nil
}

func (n *nativeEvaluator) handleSendMessages() {
	for msg := range n.out {
		if n.exited.get() {
			return
		}

		internal.Debug("Sending message: %#v", msg)
		b, err := msg.ToMsgPack()
		if err != nil {
			n.closed <- &InternalError{err: err}
			return
		}

		if err = n.client.SendMessage(b); err != nil {
			if !n.exited.get() {
				n.closed <- &InternalError{err: err}
			}
			return
		}
	}
}

func (n *nativeEvaluator) responseHandler(message []byte, userData unsafe.Pointer) {
	r := bytes.NewBuffer(message)
	dec := msgpack.NewDecoder(r)

	msg, err := msgapi.Decode(dec)
	if n.exited.get() || err == io.EOF {
		return
	}

	if err != nil {
		n.closed <- &InternalError{err: err}
		return
	}
	internal.Debug("Received message: %#v userData=%#v", msg, userData)
	n.in <- msg
}
