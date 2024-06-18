// ===----------------------------------------------------------------------===//
// Copyright © 2024 Apple Inc. and the Pkl project authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ===----------------------------------------------------------------------===//

package pkl

import (
	"context"
	"errors"
	"sync"
	"testing"

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
