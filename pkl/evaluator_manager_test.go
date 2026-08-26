//===----------------------------------------------------------------------===//
// Copyright © 2024-2026 Apple Inc. and the Pkl project authors. All rights reserved.
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

	"github.com/apple/pkl-go/pkl/internal"
	"github.com/apple/pkl-go/pkl/internal/msgapi"
	"github.com/stretchr/testify/assert"
)

type fakeEvaluatorImpl struct {
	in      chan msgapi.IncomingMessage
	out     chan msgapi.OutgoingMessage
	version string
	closed  chan error
}

func (f *fakeEvaluatorImpl) getVersion() (*internal.Semver, error) {
	if f.version == "" {
		return internal.PklVersion0_25, nil
	}
	return internal.ParseSemver(f.version)
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

func newFakeEvaluatorManager() *evaluatorManager {
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

func syncMapLen(m *sync.Map) int {
	length := 0
	m.Range(func(_, _ any) bool {
		length++
		return true
	})
	return length
}

func TestEvaluatorManager_listenContinuesAfterUnknownCreateEvaluatorResponse(t *testing.T) {
	m := newFakeEvaluatorManager()
	impl := m.impl.(*fakeEvaluatorImpl)
	impl.in = make(chan msgapi.IncomingMessage, 2)

	const knownRequestID int64 = 1
	response := make(chan *msgapi.CreateEvaluatorResponse, 1)
	m.pendingEvaluators.Store(knownRequestID, response)
	impl.in <- &msgapi.CreateEvaluatorResponse{RequestId: 2}
	impl.in <- &msgapi.CreateEvaluatorResponse{RequestId: knownRequestID}
	close(impl.in)

	go m.listen()
	select {
	case <-response:
	case <-time.After(time.Second):
		t.Fatal("dispatcher stopped after an unknown create evaluator response")
	}
}

func TestEvaluatorManager_lateCreateEvaluatorResponseDoesNotBlockDispatcher(t *testing.T) {
	m := newFakeEvaluatorManager()
	impl := m.impl.(*fakeEvaluatorImpl)
	impl.in = make(chan msgapi.IncomingMessage, 2)
	impl.out = make(chan msgapi.OutgoingMessage, 1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	evaluator, err := m.NewEvaluator(ctx)
	assert.Nil(t, evaluator)
	assert.ErrorIs(t, err, context.Canceled)

	request := (<-impl.out).(*msgapi.CreateEvaluator)
	subsequentRequestID := request.RequestId + 1
	subsequentResponse := make(chan *msgapi.CreateEvaluatorResponse, 1)
	m.pendingEvaluators.Store(subsequentRequestID, subsequentResponse)
	impl.in <- &msgapi.CreateEvaluatorResponse{RequestId: request.RequestId}
	impl.in <- &msgapi.CreateEvaluatorResponse{RequestId: subsequentRequestID}
	close(impl.in)

	go m.listen()
	select {
	case <-subsequentResponse:
	case <-time.After(time.Second):
		t.Fatal("late create evaluator response blocked the dispatcher")
	}
}

func TestEvaluatorManager_NewEvaluatorWithCanceledContext(t *testing.T) {
	m := newFakeEvaluatorManager()
	impl := m.impl.(*fakeEvaluatorImpl)
	impl.out = make(chan msgapi.OutgoingMessage, 1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	evaluator, err := m.NewEvaluator(ctx)

	assert.Nil(t, evaluator)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, syncMapLen(m.pendingEvaluators))
}

func TestEvaluator_EvaluateExpressionRawWithCanceledContext(t *testing.T) {
	m := newFakeEvaluatorManager()
	impl := m.impl.(*fakeEvaluatorImpl)
	impl.in = make(chan msgapi.IncomingMessage, 2)
	impl.out = make(chan msgapi.OutgoingMessage, 1)
	e := &evaluator{
		evaluatorId:     1,
		manager:         m,
		pendingRequests: &sync.Map{},
	}
	m.evaluators.Store(e.evaluatorId, e)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := e.EvaluateExpressionRaw(ctx, TextSource("42"), "")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, syncMapLen(e.pendingRequests))

	request := (<-impl.out).(*msgapi.Evaluate)
	subsequentRequestID := request.RequestId + 1
	subsequentResponse := make(chan *msgapi.CreateEvaluatorResponse, 1)
	m.pendingEvaluators.Store(subsequentRequestID, subsequentResponse)
	impl.in <- &msgapi.EvaluateResponse{RequestId: request.RequestId, EvaluatorId: e.evaluatorId}
	impl.in <- &msgapi.CreateEvaluatorResponse{RequestId: subsequentRequestID}
	close(impl.in)

	go m.listen()
	select {
	case <-subsequentResponse:
	case <-time.After(time.Second):
		t.Fatal("late evaluate response blocked the dispatcher")
	}
}

func TestEvaluatorManager_interrupt_NewEvaluator(t *testing.T) {
	m := newFakeEvaluatorManager()
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
	m := newFakeEvaluatorManager()
	go m.listen()
	go func() {
		_ = m.Close()
	}()
	evaluator, err := m.NewEvaluator(context.Background())
	assert.Nil(t, evaluator)
	assert.EqualError(t, err, "EvaluatorManager has been closed")
}
