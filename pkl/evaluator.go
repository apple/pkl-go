// ===----------------------------------------------------------------------===//
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
// ===----------------------------------------------------------------------===//

package pkl

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/apple/pkl-go/pkl/internal/msgapi"
)

// Evaluator is an interface for evaluating Pkl modules.
type Evaluator interface {
	// EvaluateModule evaluates the given module, and writes it to the value pointed by
	// out.
	//
	// This method is designed to work with Go modules that have been code generated from Pkl
	// sources.
	EvaluateModule(ctx context.Context, source *ModuleSource, out any) error

	// EvaluateOutputText evaluates the `output.text` property of the given module.
	EvaluateOutputText(ctx context.Context, source *ModuleSource) (string, error)

	// EvaluateOutputBytes evaluates the `output.bytes` property of the given module.
	//
	// Supported on Pkl 0.29 and higher.
	EvaluateOutputBytes(ctx context.Context, source *ModuleSource) ([]byte, error)

	// EvaluateOutputValue evaluates the `output.value` property of the given module,
	// and writes to the value pointed by out.
	EvaluateOutputValue(ctx context.Context, source *ModuleSource, out any) error

	// EvaluateOutputFiles evaluates the `output.files` property of the given module, giving the text of each file.
	EvaluateOutputFiles(ctx context.Context, source *ModuleSource) (map[string]string, error)

	// EvaluateOutputFilesBytes evaluates the `output.files` property of the given module, giving the bytes of each file.
	//
	// Supported on Pkl 0.29 and higher.
	EvaluateOutputFilesBytes(ctx context.Context, source *ModuleSource) (map[string][]byte, error)

	// EvaluateExpression evaluates the provided expression on the given module source, and writes
	// the result into the value pointed by out.
	EvaluateExpression(ctx context.Context, source *ModuleSource, expr string, out any) error

	// EvaluateExpressionRaw evaluates the provided module, and returns the underlying value's raw
	// bytes.
	//
	// This is a low level API.
	EvaluateExpressionRaw(ctx context.Context, source *ModuleSource, expr string) ([]byte, error)

	// Close closes the evaluator and releases any underlying resources.
	Close() error

	// Closed tells if this evaluator is closed.
	Closed() bool
}

type evaluator struct {
	evaluatorId     int64
	logger          Logger
	manager         *evaluatorManager
	pendingRequests *sync.Map
	closed          bool
	resourceReaders []ResourceReader
	moduleReaders   []ModuleReader
}

var _ Evaluator = (*evaluator)(nil)

func (e *evaluator) EvaluateModule(ctx context.Context, source *ModuleSource, out any) error {
	return e.EvaluateExpression(ctx, source, "", out)
}

func (e *evaluator) EvaluateOutputText(ctx context.Context, source *ModuleSource) (string, error) {
	var out string
	err := e.EvaluateExpression(ctx, source, "output.text", &out)
	return out, err
}

func (e *evaluator) EvaluateOutputBytes(ctx context.Context, source *ModuleSource) ([]byte, error) {
	var out []byte
	err := e.EvaluateExpression(ctx, source, "output.bytes", &out)
	return out, err
}

func (e *evaluator) EvaluateOutputValue(ctx context.Context, source *ModuleSource, out any) error {
	return e.EvaluateExpression(ctx, source, "output.value", out)
}

func (e *evaluator) EvaluateOutputFiles(ctx context.Context, source *ModuleSource) (map[string]string, error) {
	var out map[string]string
	err := e.EvaluateExpression(ctx, source, "output.files?.toMap()?.mapValues((_, it) -> it.text) ?? Map()", &out)
	return out, err
}

func (e *evaluator) EvaluateOutputFilesBytes(ctx context.Context, source *ModuleSource) (map[string][]byte, error) {
	var out map[string][]byte
	err := e.EvaluateExpression(ctx, source, "output.files?.toMap()?.mapValues((_, it) -> it.bytes) ?? Map()", &out)
	return out, err
}

func (e *evaluator) EvaluateExpression(ctx context.Context, source *ModuleSource, expr string, out any) error {
	bytes, err := e.EvaluateExpressionRaw(ctx, source, expr)
	if err != nil {
		return err
	}
	return Unmarshal(bytes, out)
}

func (e *evaluator) EvaluateExpressionRaw(ctx context.Context, source *ModuleSource, expr string) ([]byte, error) {
	if e.Closed() {
		return nil, fmt.Errorf("evaluator is closed")
	}
	requestId := random.Int63()
	ch := make(chan *msgapi.EvaluateResponse)
	e.pendingRequests.Store(requestId, ch)
	interrupted, nevermind := e.manager.interrupted(e.evaluatorId)
	defer nevermind()
	e.manager.impl.outChan() <- &msgapi.Evaluate{
		RequestId:   requestId,
		ModuleUri:   source.Uri.String(),
		ModuleText:  source.Contents,
		Expr:        expr,
		EvaluatorId: e.evaluatorId,
	}
	select {
	case <-ctx.Done():
		return nil, nil
	case err := <-interrupted:
		return nil, err
	case resp := <-ch:
		if resp.Error != "" {
			return nil, &EvalError{ErrorOutput: resp.Error}
		}
		return resp.Result, nil
	}
}

func (e *evaluator) Close() error {
	if e.closed {
		return nil
	}
	e.manager.closeEvaluator(e)
	return nil
}

func (e *evaluator) Closed() bool {
	return e.closed
}

func (e *evaluator) handleEvaluateResponse(resp *msgapi.EvaluateResponse) {
	c, exists := e.pendingRequests.Load(resp.RequestId)
	if !exists {
		log.Default().Printf("warn: received a message for an unknown request id: %d", resp.RequestId)
		return
	}
	ch := c.(chan *msgapi.EvaluateResponse)
	ch <- resp
	close(ch)
	e.pendingRequests.Delete(resp.RequestId)
}

func (e *evaluator) handleLog(resp *msgapi.Log) {
	switch resp.Level {
	case 0:
		e.logger.Trace(resp.Message, resp.FrameUri)
	case 1:
		e.logger.Warn(resp.Message, resp.FrameUri)
	default:
		// log level beyond 1 is impossible
		panic(fmt.Sprintf("unknown log level: %d", resp.Level))
	}
}

func (e *evaluator) handleReadResource(msg *msgapi.ReadResource) {
	response := &msgapi.ReadResourceResponse{EvaluatorId: e.evaluatorId, RequestId: msg.RequestId}
	u, err := url.Parse(msg.Uri)
	if err != nil {
		response.Error = fmt.Errorf("internal error: failed to parse resource url: %w", err).Error()
		e.manager.impl.outChan() <- response
		return
	}
	reader := e.findResourceReader(u.Scheme)
	if reader == nil {
		response.Error = fmt.Sprintf("No resource reader found for scheme `%s`", u.Scheme)
		e.manager.impl.outChan() <- response
		return
	}
	contents, err := reader.Read(*u)
	response.Contents = contents
	if err != nil {
		response.Error = err.Error()
	}
	e.manager.impl.outChan() <- response
}

func (e *evaluator) handleReadModule(msg *msgapi.ReadModule) {
	response := &msgapi.ReadModuleResponse{EvaluatorId: e.evaluatorId, RequestId: msg.RequestId}
	u, err := url.Parse(msg.Uri)
	if err != nil {
		response.Error = fmt.Errorf("internal error: failed to parse resource url: %w", err).Error()
		e.manager.impl.outChan() <- response
		return
	}
	reader := e.findModuleReader(u.Scheme)
	if reader == nil {
		response.Error = fmt.Sprintf("No module reader found for scheme `%s`", u.Scheme)
		e.manager.impl.outChan() <- response
		return
	}
	response.Contents, err = reader.Read(*u)
	if err != nil {
		response.Error = err.Error()
	}
	e.manager.impl.outChan() <- response
}

func (e *evaluator) handleListResources(msg *msgapi.ListResources) {
	response := &msgapi.ListResourcesResponse{
		EvaluatorId: e.evaluatorId,
		RequestId:   msg.RequestId,
	}

	u, err := url.Parse(msg.Uri)
	if err != nil {
		response.Error = fmt.Errorf("internal error: failed to parse resource url: %w", err).Error()
		e.manager.impl.outChan() <- response
		return
	}

	reader := e.findResourceReader(u.Scheme)
	if reader == nil {
		response.Error = fmt.Sprintf("No resource reader found for scheme `%s`", u.Scheme)
		e.manager.impl.outChan() <- response
		return
	}

	pathElements, err := reader.ListElements(*u)
	if err != nil {
		response.Error = err.Error()
	} else {
		response.PathElements = make([]*msgapi.PathElement, len(pathElements))
		for i, pe := range pathElements {
			response.PathElements[i] = &msgapi.PathElement{
				Name:        pe.Name(),
				IsDirectory: pe.IsDirectory(),
			}
		}
	}

	e.manager.impl.outChan() <- response
}

func (e *evaluator) handleListModules(msg *msgapi.ListModules) {
	response := &msgapi.ListModulesResponse{EvaluatorId: e.evaluatorId, RequestId: msg.RequestId}
	u, err := url.Parse(msg.Uri)
	if err != nil {
		response.Error = fmt.Errorf("internal error: failed to parse resource url: %w", err).Error()
		e.manager.impl.outChan() <- response
		return
	}
	reader := e.findModuleReader(u.Scheme)
	if reader == nil {
		response.Error = fmt.Sprintf("No module reader found for scheme `%s`", u.Scheme)
		e.manager.impl.outChan() <- response
		return
	}
	pathElements, err := reader.ListElements(*u)
	if err != nil {
		response.Error = err.Error()
	} else {
		for _, pathElement := range pathElements {
			response.PathElements = append(response.PathElements, &msgapi.PathElement{
				Name:        pathElement.Name(),
				IsDirectory: pathElement.IsDirectory(),
			})
		}
	}
	e.manager.impl.outChan() <- response
}

func (e *evaluator) findModuleReader(scheme string) ModuleReader {
	for _, r := range e.moduleReaders {
		if r.Scheme() == scheme {
			return r
		}
	}
	return nil
}

func (e *evaluator) findResourceReader(scheme string) ResourceReader {
	for _, r := range e.resourceReaders {
		if r.Scheme() == scheme {
			return r
		}
	}
	return nil
}

type simpleEvaluator struct {
	Evaluator
	manager EvaluatorManager
}

var _ Evaluator = (*simpleEvaluator)(nil)

func (rcv *simpleEvaluator) Close() error {
	return rcv.manager.Close()
}
