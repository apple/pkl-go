//===----------------------------------------------------------------------===//
// Copyright © 2024-2025 Apple Inc. and the Pkl project authors. All rights reserved.
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

package pkl

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/apple/pkl-go/pkl/internal/msgapi"
	"github.com/stretchr/testify/assert"
)

type fakeEvaluatorImpl struct {
	in      chan msgapi.IncomingMessage
	out     chan msgapi.OutgoingMessage
	version string
	closed  chan error
}

func (f *fakeEvaluatorImpl) getVersion() (*semver, error) {
	if f.version == "" {
		return pklVersion0_25, nil
	}
	return parseSemver(f.version)
}

func (f *fakeEvaluatorImpl) init() error {
	return nil
}

func (f *fakeEvaluatorImpl) deinit() error {
	return nil
}

func (f *fakeEvaluatorImpl) inChan() chan msgapi.IncomingMessage {
	return f.in
}

func (f *fakeEvaluatorImpl) outChan() chan msgapi.OutgoingMessage {
	return f.out
}

func (f *fakeEvaluatorImpl) closedChan() chan error {
	return f.closed
}

var _ evaluatorManagerImpl = (*fakeEvaluatorImpl)(nil)

func newFakeEvalautorManager() *evaluatorManager {
	return &evaluatorManager{
		impl: &fakeEvaluatorImpl{
			in:     make(chan msgapi.IncomingMessage),
			out:    make(chan msgapi.OutgoingMessage),
			closed: make(chan error),
		},
		interrupts:        &sync.Map{},
		evaluators:        &sync.Map{},
		pendingEvaluators: &sync.Map{},
	}
}

func TestEvaluatorManager_interrupt_NewEvaluator(t *testing.T) {
	m := newFakeEvalautorManager()
	defer assert.NoError(t, m.Close())
	go m.listen()
	go func() {
		m.interrupt(errors.New("test interruption"))
	}()
	evaluator, err := m.NewEvaluator(context.Background())
	assert.Nil(t, evaluator)
	assert.Error(t, err, "test interruption")
}

func TestEvaluatorManager_interrupt_Close(t *testing.T) {
	m := newFakeEvalautorManager()
	go m.listen()
	go func() {
		_ = m.Close()
	}()
	evaluator, err := m.NewEvaluator(context.Background())
	assert.Nil(t, evaluator)
	assert.Nil(t, err)
}

// TestEvaluator_EvaluateExpressionRaw_CtxCancel verifies that a cancelled context
// surfaces as ctx.Err() (not a swallowed (nil, nil)), that the pendingRequests
// entry is cleaned up, and that a response which arrives after the caller has
// already given up does not block the manager's shared listen() goroutine -
// which previously would deadlock every future evaluation on this evaluator.
func TestEvaluator_EvaluateExpressionRaw_CtxCancel(t *testing.T) {
	m := newFakeEvalautorManager()
	go m.listen()

	requestIdCh := make(chan int64, 1)
	go func() {
		msg := <-m.impl.outChan()
		if ev, ok := msg.(*msgapi.Evaluate); ok {
			requestIdCh <- ev.RequestId
		}
	}()

	ev := &evaluator{
		evaluatorId:     1,
		manager:         m,
		pendingRequests: &sync.Map{},
	}
	m.evaluators.Store(int64(1), ev)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := ev.EvaluateExpressionRaw(ctx, TextSource("foo"), "output.text")
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.Canceled)

	var requestId int64
	select {
	case requestId = <-requestIdCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the Evaluate request to be sent")
	}

	_, exists := ev.pendingRequests.Load(requestId)
	assert.False(t, exists, "pendingRequests entry should be cleaned up after ctx cancellation")

	// Simulate the pkl process finishing the evaluation after the caller already
	// gave up. Before this fix, handleEvaluateResponse would block forever
	// sending on the abandoned, unbuffered response channel, wedging listen()
	// for every other evaluation on this evaluator. A second, unrelated
	// response is sent afterwards as a sentinel: since the fake inChan is
	// itself unbuffered, listen() can only receive it once it has returned
	// from handling the first (proving it didn't get stuck inside it).
	done := make(chan struct{})
	go func() {
		m.impl.inChan() <- &msgapi.EvaluateResponse{
			RequestId:   requestId,
			EvaluatorId: 1,
			Result:      []byte("late response"),
		}
		m.impl.inChan() <- &msgapi.EvaluateResponse{
			RequestId:   requestId + 1, // unknown request id: handled as a harmless no-op
			EvaluatorId: 1,
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("listen() appears deadlocked processing a response for an abandoned request")
	}
}
