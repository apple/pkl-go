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
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/apple/pkl-go/pkl/internal"
	"github.com/apple/pkl-go/pkl/internal/msgapi"
	"github.com/vmihailenco/msgpack/v5"
)

// ExternalReaderClient is an interface for implementing [external readers](https://pkl-lang.org/main/current/language-reference/index.html#external-readers).
type ExternalReaderClient interface {
	// Run starts the ExternalReaderClient and blocks until the reader is closed by [ExternalReaderClient.Close] or the Pkl evaluator.
	Run() error

	// Close disconnects the ExternalReaderClient from the Pkl evaluator and cleans up any resources.
	Close()
}

// ExternalReaderClientOptions is the set of options available to control ExternalReaderClients.
type ExternalReaderClientOptions struct {
	// RequestReader is the interface used to read messages from the Pkl evaluator. If omitted, os.Stdin will be used.
	RequestReader io.Reader

	// ResponseWriter is the interface used to send messages to the Pkl evaluator. If omitted, os.Stdout will be used.
	ResponseWriter io.Writer

	// ResourceReaders are the resource readers to be used by the evaluator.
	ResourceReaders []ResourceReader

	// ModuleReaders are the set of custom module readers to be used by the evaluator.
	ModuleReaders []ModuleReader
}

// WithExternalClientResourceReader adds an additional [ResourceReader] to the ExternalReaderClient.
var WithExternalClientResourceReader = func(reader ResourceReader) func(*ExternalReaderClientOptions) {
	return func(options *ExternalReaderClientOptions) {
		options.ResourceReaders = append(options.ResourceReaders, reader)
	}
}

// WithExternalClientModuleReader adds an additional [ModuleReader] to the ExternalReaderClient.
var WithExternalClientModuleReader = func(reader ModuleReader) func(*ExternalReaderClientOptions) {
	return func(options *ExternalReaderClientOptions) {
		options.ModuleReaders = append(options.ModuleReaders, reader)
	}
}

// WithExternalClientStreams sets the input and output interfaces the ExternalReaderClient will use to communicate with the Pkl evaluator.
var WithExternalClientStreams = func(requestReader io.Reader, responseWriter io.Writer) func(*ExternalReaderClientOptions) {
	return func(options *ExternalReaderClientOptions) {
		options.RequestReader = requestReader
		options.ResponseWriter = responseWriter
	}
}

// NewExternalReaderClient creates a new ExternalReaderClient.
func NewExternalReaderClient(opts ...func(options *ExternalReaderClientOptions)) (ExternalReaderClient, error) {
	o := ExternalReaderClientOptions{}
	for _, f := range opts {
		f(&o)
	}

	if o.RequestReader == nil {
		o.RequestReader = os.Stdin
	}
	if o.ResponseWriter == nil {
		o.ResponseWriter = os.Stdout
	}

	return &externalReaderClient{
		ExternalReaderClientOptions: o,
		in:                          make(chan msgapi.IncomingMessage),
		out:                         make(chan msgapi.OutgoingMessage),
		closed:                      make(chan error),
	}, nil
}

type externalReaderClient struct {
	ExternalReaderClientOptions
	in     chan msgapi.IncomingMessage
	out    chan msgapi.OutgoingMessage
	closed chan error
	exited atomicBool
}

var _ ExternalReaderClient = (*externalReaderClient)(nil)

func (r *externalReaderClient) Run() error {
	internal.Debug("Starting external reader client")
	for _, reader := range r.ModuleReaders {
		internal.Debug("Registered module reader of type %T for scheme %q", reader, reader.Scheme())
	}
	for _, reader := range r.ResourceReaders {
		internal.Debug("Registered resource reader of type %T for scheme %q", reader, reader.Scheme())
	}

	go r.readIncomingMessages()
	go r.handleSendMessages()
	go r.listen()

	return <-r.closed
}

func (r *externalReaderClient) Close() {
	r.exited.set(true)
	close(r.in)
	close(r.out)
	close(r.closed)
}

func (r *externalReaderClient) readIncomingMessages() {
	dec := msgpack.NewDecoder(r.RequestReader)
	for {
		msg, err := msgapi.Decode(dec)
		if r.exited.get() || err == io.EOF {
			break
		}
		if err != nil {
			r.closed <- &InternalError{err: err}
			return
		}
		internal.Debug("Received message: %#v", msg)
		r.in <- msg
	}
}

func (r *externalReaderClient) handleSendMessages() {
	for msg := range r.out {
		internal.Debug("Sending message: %#v", msg)
		b, err := msg.ToMsgPack()
		if r.exited.get() {
			return
		}
		if err != nil {
			r.closed <- &InternalError{err: err}
			return
		}
		if _, err = r.ResponseWriter.Write(b); err != nil {
			if !r.exited.get() {
				r.closed <- &InternalError{err: err}
			}
			return
		}
	}
}

func (r *externalReaderClient) listen() {
	for msg := range r.in {
		switch msg := msg.(type) {
		case *msgapi.InitializeModuleReader:
			r.handleInitializeModuleReader(msg)
		case *msgapi.InitializeResourceReader:
			r.handleInitializeResourceReader(msg)
		case *msgapi.ReadResource:
			r.handleReadResource(msg)
		case *msgapi.ReadModule:
			r.handleReadModule(msg)
		case *msgapi.ListResources:
			r.handleListResources(msg)
		case *msgapi.ListModules:
			r.handleListModules(msg)
		case *msgapi.CloseExternalProcess:
			r.Close()
		}
	}
}

func (r *externalReaderClient) handleInitializeModuleReader(msg *msgapi.InitializeModuleReader) {
	for _, reader := range r.ModuleReaders {
		if reader.Scheme() == msg.Scheme {
			r.out <- &msgapi.InitializeModuleReaderResponse{
				RequestId: msg.RequestId,
				Spec: &msgapi.ModuleReader{
					Scheme:              reader.Scheme(),
					IsGlobbable:         reader.IsGlobbable(),
					HasHierarchicalUris: reader.HasHierarchicalUris(),
					IsLocal:             reader.IsLocal(),
				},
			}
			return
		}
	}
	r.out <- &msgapi.InitializeModuleReaderResponse{
		RequestId: msg.RequestId,
	}
}

func (r *externalReaderClient) handleInitializeResourceReader(msg *msgapi.InitializeResourceReader) {
	for _, reader := range r.ResourceReaders {
		if reader.Scheme() == msg.Scheme {
			r.out <- &msgapi.InitializeResourceReaderResponse{
				RequestId: msg.RequestId,
				Spec: &msgapi.ResourceReader{
					Scheme:              reader.Scheme(),
					IsGlobbable:         reader.IsGlobbable(),
					HasHierarchicalUris: reader.HasHierarchicalUris(),
				},
			}
			return
		}
	}
	r.out <- &msgapi.InitializeResourceReaderResponse{
		RequestId: msg.RequestId,
	}
}

func (r *externalReaderClient) handleReadResource(msg *msgapi.ReadResource) {
	response := &msgapi.ReadResourceResponse{EvaluatorId: msg.EvaluatorId, RequestId: msg.RequestId}
	u, err := url.Parse(msg.Uri)
	if err != nil {
		response.Error = fmt.Errorf("internal error: failed to parse resource url: %w", err).Error()
		r.out <- response
		return
	}
	var reader ResourceReader
	for _, r := range r.ResourceReaders {
		if r.Scheme() == u.Scheme {
			reader = r
			break
		}
	}
	if reader == nil {
		response.Error = fmt.Sprintf("No resource reader found for scheme `%s`", u.Scheme)
		r.out <- response
		return
	}
	contents, err := reader.Read(*u)
	response.Contents = contents
	if err != nil {
		response.Error = err.Error()
	}
	r.out <- response
}

func (r *externalReaderClient) handleReadModule(msg *msgapi.ReadModule) {
	response := &msgapi.ReadModuleResponse{EvaluatorId: msg.EvaluatorId, RequestId: msg.RequestId}
	u, err := url.Parse(msg.Uri)
	if err != nil {
		response.Error = fmt.Errorf("internal error: failed to parse resource url: %w", err).Error()
		r.out <- response
		return
	}
	var reader ModuleReader
	for _, r := range r.ModuleReaders {
		if r.Scheme() == u.Scheme {
			reader = r
			break
		}
	}
	if reader == nil {
		response.Error = fmt.Sprintf("No module reader found for scheme `%s`", u.Scheme)
		r.out <- response
		return
	}
	response.Contents, err = reader.Read(*u)
	if err != nil {
		response.Error = err.Error()
	}
	r.out <- response
}

func (r *externalReaderClient) handleListResources(msg *msgapi.ListResources) {
	response := &msgapi.ListResourcesResponse{EvaluatorId: msg.EvaluatorId, RequestId: msg.RequestId}
	u, err := url.Parse(msg.Uri)
	if err != nil {
		response.Error = fmt.Errorf("internal error: failed to parse resource url: %w", err).Error()
		r.out <- response
		return
	}
	var reader ResourceReader
	for _, r := range r.ResourceReaders {
		if r.Scheme() == u.Scheme {
			reader = r
			break
		}
	}
	if reader == nil {
		response.Error = fmt.Sprintf("No resource reader found for scheme `%s`", u.Scheme)
		r.out <- response
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
	r.out <- response
}

func (r *externalReaderClient) handleListModules(msg *msgapi.ListModules) {
	response := &msgapi.ListModulesResponse{EvaluatorId: msg.EvaluatorId, RequestId: msg.RequestId}
	u, err := url.Parse(msg.Uri)
	if err != nil {
		response.Error = fmt.Errorf("internal error: failed to parse resource url: %w", err).Error()
		r.out <- response
		return
	}
	var reader ModuleReader
	for _, r := range r.ModuleReaders {
		if r.Scheme() == u.Scheme {
			reader = r
			break
		}
	}
	if reader == nil {
		response.Error = fmt.Sprintf("No module reader found for scheme `%s`", u.Scheme)
		r.out <- response
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
	r.out <- response
}
