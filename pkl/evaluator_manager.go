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
	"log"
	"net/url"
	"sync"

	"github.com/apple/pkl-go/pkl/internal"
	"github.com/apple/pkl-go/pkl/internal/msgapi"
)

var empty = struct{}{}

// EvaluatorManager provides a way to minimize the overhead of multiple evaluators.
// For example, if calling into Pkl as a child process, using the manager will share the same
// process for all created evaluators. In contrast, constructing multiple evaluators through
// NewEvaluator will spawn one process per evaluator.
type EvaluatorManager interface {
	// Close closes the evaluator manager and all of its evaluators.
	//
	// If running Pkl as a child process, closes all evaluators as well as the child process.
	// If calling into Pkl through the C API, close all existing evaluators.
	Close() error

	// GetVersion returns the version of Pkl backing this evaluator manager.
	GetVersion() (string, error)

	// NewEvaluator constructs an evaluator instance.
	//
	// If calling into Pkl as a child process, the first time NewEvaluator is called, this will
	// start the child process.
	NewEvaluator(ctx context.Context, opts ...func(options *EvaluatorOptions)) (Evaluator, error)

	// NewProjectEvaluator is an easy way to create an evaluator whose project directory is determined
	// by `projectBaseUrl`.
	// It loads the project from the `PklProject` and `PklProject.deps.json` files within `projectBaseUrl`.
	//
	// It is similar to running the `pkl eval` or `pkl test` CLI command with a set `--project-dir`.
	//
	// When using project dependencies, they must first be resolved using the `pkl project resolve`
	// CLI command.
	NewProjectEvaluator(ctx context.Context, projectBaseUrl *url.URL, opts ...func(options *EvaluatorOptions)) (Evaluator, error)
}

type evaluatorManager struct {
	impl              evaluatorManagerImpl
	interrupts        *sync.Map
	evaluators        *sync.Map
	pendingEvaluators *sync.Map
	closed            atomicBool
	newEvaluatorMutex sync.Mutex
	initialized       bool
}

// evaluatorManagerImpl is the underlying implementation of the manager. It defines the logic
// behind setup and teardown routines, and provides channels for incoming/outgoing messages and
// out-of-band closes.
type evaluatorManagerImpl interface {
	init() error
	deinit() error
	inChan() chan msgapi.IncomingMessage
	outChan() chan msgapi.OutgoingMessage
	closedChan() chan error
	getVersion() (*internal.Semver, error)
}

var _ EvaluatorManager = (*evaluatorManager)(nil)

func (m *evaluatorManager) NewEvaluator(ctx context.Context, opts ...func(options *EvaluatorOptions)) (Evaluator, error) {
	// Prevent concurrent calls to NewEvaluator because only the first call should call the `init` routine.
	m.newEvaluatorMutex.Lock()
	defer m.newEvaluatorMutex.Unlock()
	if m.closed.get() {
		return nil, errors.New("EvaluatorManager has been closed")
	}
	if !m.initialized {
		if err := m.init(); err != nil {
			return nil, err
		}
		m.initialized = true
	}
	version, err := m.getVersion()
	if err != nil {
		return nil, err
	}
	o, err := buildEvaluatorOptions(version, opts...)
	if err != nil {
		return nil, err
	}
	var newEvaluatorRequest msgapi.OutgoingMessage
	requestId := random.Int63()
	msg := o.toMessage()
	msg.RequestId = requestId
	newEvaluatorRequest = msg
	ch := make(chan *msgapi.CreateEvaluatorResponse)
	m.pendingEvaluators.Store(requestId, ch)
	interrupt, nevermind := m.interrupted(0)
	defer nevermind()
	go func() {
		m.impl.outChan() <- newEvaluatorRequest
	}()
	// sanity check: it's possible that the evaluator has been closed at this point.
	if m.closed.get() {
		return nil, nil
	}
	select {
	case <-ctx.Done():
		return nil, nil
	case err := <-interrupt:
		return nil, err
	case resp := <-ch:
		if resp.Error != "" {
			return nil, errors.New(resp.Error)
		}
		ev := &evaluator{
			evaluatorId:     resp.EvaluatorId,
			logger:          o.Logger,
			manager:         m,
			pendingRequests: &sync.Map{},
			resourceReaders: o.ResourceReaders,
			moduleReaders:   o.ModuleReaders,
		}
		m.evaluators.Store(resp.EvaluatorId, ev)
		return ev, nil
	}
}

func (m *evaluatorManager) NewProjectEvaluator(ctx context.Context, projectBaseUrl *url.URL, opts ...func(options *EvaluatorOptions)) (Evaluator, error) {
	projectEvaluator, err := NewEvaluator(ctx, opts...)
	if err != nil {
		return nil, err
	}
	projectSource := projectBaseUrl.JoinPath("PklProject")
	project, err := LoadProjectFromEvaluator(ctx, projectEvaluator, &ModuleSource{Uri: projectSource})
	if err != nil {
		return nil, err
	}
	newOpts := []func(options *EvaluatorOptions){
		WithProject(project),
	}
	newOpts = append(newOpts, opts...)
	return NewEvaluator(ctx, newOpts...)
}

func (m *evaluatorManager) getVersion() (*internal.Semver, error) {
	return m.impl.getVersion()
}

func (m *evaluatorManager) GetVersion() (string, error) {
	version, err := m.getVersion()
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

func (m *evaluatorManager) Close() error {
	return m.closeErr(nil)
}

func (m *evaluatorManager) getEvaluator(evaluatorId int64) *evaluator {
	v, exists := m.evaluators.Load(evaluatorId)
	if !exists {
		log.Default().Printf("warn: received a message for an unknown evaluator id: %d", evaluatorId)
		return nil
	}
	return v.(*evaluator)
}

func (m *evaluatorManager) listen() {
	for msg := range m.impl.inChan() {
		switch msg := msg.(type) {
		case *msgapi.EvaluateResponse:
			ev := m.getEvaluator(msg.EvaluatorId)
			if ev == nil {
				return
			}
			ev.handleEvaluateResponse(msg)
		case *msgapi.Log:
			ev := m.getEvaluator(msg.EvaluatorId)
			if ev == nil {
				return
			}
			ev.handleLog(msg)
		case *msgapi.ReadResource:
			ev := m.getEvaluator(msg.EvaluatorId)
			if ev == nil {
				return
			}
			ev.handleReadResource(msg)
		case *msgapi.ReadModule:
			ev := m.getEvaluator(msg.EvaluatorId)
			if ev == nil {
				return
			}
			ev.handleReadModule(msg)
		case *msgapi.ListResources:
			ev := m.getEvaluator(msg.EvaluatorId)
			if ev == nil {
				return
			}
			ev.handleListResources(msg)
		case *msgapi.ListModules:
			ev := m.getEvaluator(msg.EvaluatorId)
			if ev == nil {
				return
			}
			ev.handleListModules(msg)
		case *msgapi.CreateEvaluatorResponse:
			ch, exists := m.pendingEvaluators.Load(msg.RequestId)
			if !exists {
				log.Default().Printf("warn: received a message for an unknown request id: %d", msg.RequestId)
				return
			}
			cch := ch.(chan *msgapi.CreateEvaluatorResponse)
			cch <- msg
			close(cch)
			m.pendingEvaluators.Delete(msg.RequestId)
		}
	}
}

// listenForImplClose handles sudden interruption of the Evaluator, for example, if
// the pkl child process suddenly exits.
//
// This method will also be called if EvaluatorManager.Close is explicitly called. But it is safe
// to call `closeErr` multiple times.
func (m *evaluatorManager) listenForImplClose() {
	err := <-m.impl.closedChan()
	_ = m.closeErr(err)
}

// interrupted creates a channel that gets published to when an interruption happens, and also
// a function to clean up the channel.
//
// evaluatorId is the optional ID of an evaluator.
//
// Possible reason for interruptions:
//   - The underlying pkl process died
//   - The EvaluatorManager was closed
//   - The Evaluator was closed
func (m *evaluatorManager) interrupted(evaluatorId int64) (chan error, func()) {
	ch := make(chan error)
	m.interrupts.Store(ch, evaluatorId)
	return ch, func() {
		m.interrupts.Delete(ch)
	}
}

// closeEvaluator closes the provided evaluator.
func (m *evaluatorManager) closeEvaluator(ev *evaluator) {
	// if the manager itself is closed, there's nothing to do.
	if m.closed.get() {
		return
	}
	m.impl.outChan() <- &msgapi.CloseEvaluator{EvaluatorId: ev.evaluatorId}
	m.evaluators.Delete(ev.evaluatorId)
	ev.closed.set(true)
	m.interrupts.Range(func(key, value any) bool {
		if value.(int64) == ev.evaluatorId {
			key.(chan error) <- nil
		}
		return true
	})
}

func (m *evaluatorManager) interrupt(err error) {
	m.interrupts.Range(func(ch, _ any) bool {
		ch.(chan error) <- err
		return true
	})
}

func (m *evaluatorManager) init() error {
	return m.impl.init()
}

func (m *evaluatorManager) closeErr(e error) error {
	if m.closed.get() {
		return nil
	}
	var err error
	m.interrupt(e)
	m.evaluators.Range(func(evaluatorId, v any) bool {
		ev := v.(*evaluator)
		// if an error occurs, still try to keep closing.
		if cerr := ev.Close(); cerr != nil {
			err = cerr
		}
		return true
	})
	m.closed.set(true)
	derr := m.impl.deinit()
	if err != nil {
		return err
	}
	return derr
}
