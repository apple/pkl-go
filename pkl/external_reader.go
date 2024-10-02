package pkl

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/apple/pkl-go/pkl/internal"
	"github.com/apple/pkl-go/pkl/internal/msgapi"
	"github.com/vmihailenco/msgpack/v5"
)

type ExternalReaderRuntime interface {
	Run() error
	Close()
}

type ExternalReaderRuntimeOptions struct {
	// Reader to receive requests. If omitted, os.Stdin will be used
	RequestReader io.Reader

	// Writer to publish responses. If omitted, os.Stdout will be used
	ResponseWriter io.Writer

	// ResourceReaders are the resource readers to be used by the evaluator.
	ResourceReaders []ResourceReader

	// ModuleReaders are the set of custom module readers to be used by the evaluator.
	ModuleReaders []ModuleReader
}

func NewExternalReaderRuntime(ctx context.Context, opts ...func(options *ExternalReaderRuntimeOptions)) (ExternalReaderRuntime, error) {
	o := ExternalReaderRuntimeOptions{}
	for _, f := range opts {
		f(&o)
	}

	if o.RequestReader == nil {
		o.RequestReader = os.Stdin
	}
	if o.ResponseWriter == nil {
		o.ResponseWriter = os.Stdout
	}

	return &externalReaderRuntime{
		ExternalReaderRuntimeOptions: o,
		in:                           make(chan msgapi.IncomingMessage),
		out:                          make(chan msgapi.OutgoingMessage),
		closed:                       make(chan error),
	}, nil
}

type externalReaderRuntime struct {
	ExternalReaderRuntimeOptions
	in     chan msgapi.IncomingMessage
	out    chan msgapi.OutgoingMessage
	closed chan error
	exited atomicBool
}

var _ ExternalReaderRuntime = (*externalReaderRuntime)(nil)

func (r *externalReaderRuntime) Run() error {
	// XXX Does it mae sense to check if RequestReader/Write are TTYs and throw an error if so?

	internal.Debug("Starting external reader runtime")
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

func (r *externalReaderRuntime) Close() {
	r.exited.set(true)
	close(r.in)
	close(r.out)
	close(r.closed)
}

func (r *externalReaderRuntime) readIncomingMessages() {
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

func (r *externalReaderRuntime) handleSendMessages() {
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

func (r *externalReaderRuntime) listen() {
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

func (r *externalReaderRuntime) handleInitializeModuleReader(msg *msgapi.InitializeModuleReader) {
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

func (r *externalReaderRuntime) handleInitializeResourceReader(msg *msgapi.InitializeResourceReader) {
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

func (r *externalReaderRuntime) handleReadResource(msg *msgapi.ReadResource) {
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

func (r *externalReaderRuntime) handleReadModule(msg *msgapi.ReadModule) {
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

func (r *externalReaderRuntime) handleListResources(msg *msgapi.ListResources) {
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

func (r *externalReaderRuntime) handleListModules(msg *msgapi.ListModules) {
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
