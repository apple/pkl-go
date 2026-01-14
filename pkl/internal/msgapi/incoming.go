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

package msgapi

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

type IncomingMessage interface {
	incomingMessage()
}

type incomingMessageImpl struct{}

func (r incomingMessageImpl) incomingMessage() {}

var (
	_ IncomingMessage = (*CreateEvaluatorResponse)(nil)
	_ IncomingMessage = (*EvaluateResponse)(nil)
	_ IncomingMessage = (*ReadResource)(nil)
	_ IncomingMessage = (*ReadModule)(nil)
	_ IncomingMessage = (*Log)(nil)
	_ IncomingMessage = (*ListResources)(nil)
	_ IncomingMessage = (*ListModules)(nil)
	_ IncomingMessage = (*CloseExternalProcess)(nil)
)

type CreateEvaluatorResponse struct {
	incomingMessageImpl

	Error string `msgpack:"error"`

	RequestId   int64 `msgpack:"requestId"`
	EvaluatorId int64 `msgpack:"evaluatorId"`
}

type EvaluateResponse struct {
	incomingMessageImpl

	Error  string `msgpack:"error"`
	Result []byte `msgpack:"result"`

	RequestId   int64 `msgpack:"requestId"`
	EvaluatorId int64 `msgpack:"evaluatorId"`
}

type ReadResource struct {
	incomingMessageImpl

	Uri string `msgpack:"uri"`

	RequestId   int64 `msgpack:"requestId"`
	EvaluatorId int64 `msgpack:"evaluatorId"`
}

type ReadModule struct {
	incomingMessageImpl

	Uri string `msgpack:"uri"`

	RequestId   int64 `msgpack:"requestId"`
	EvaluatorId int64 `msgpack:"evaluatorId"`
}

type Log struct {
	incomingMessageImpl

	Message  string `msgpack:"message"`
	FrameUri string `msgpack:"frameUri"`

	EvaluatorId int64 `msgpack:"evaluatorId"`
	Level       int   `msgpack:"level"`
}

type ListResources struct {
	incomingMessageImpl

	Uri string `msgpack:"uri"`

	RequestId   int64 `msgpack:"requestId"`
	EvaluatorId int64 `msgpack:"evaluatorId"`
}

type ListModules struct {
	incomingMessageImpl

	Uri string `msgpack:"uri"`

	RequestId   int64 `msgpack:"requestId"`
	EvaluatorId int64 `msgpack:"evaluatorId"`
}

type InitializeModuleReader struct {
	incomingMessageImpl

	Scheme string `msgpack:"scheme"`

	RequestId int64 `msgpack:"requestId"`
}

type InitializeResourceReader struct {
	incomingMessageImpl

	Scheme string `msgpack:"scheme"`

	RequestId int64 `msgpack:"requestId"`
}

type CloseExternalProcess struct {
	incomingMessageImpl
}

func Decode(decoder *msgpack.Decoder) (IncomingMessage, error) {
	_, err := decoder.DecodeArrayLen()
	if err != nil {
		return nil, err
	}
	c, err := decoder.DecodeInt()
	if err != nil {
		return nil, err
	}
	switch c {
	case codeEvaluateResponse:
		var resp EvaluateResponse
		return &resp, decoder.Decode(&resp)
	case codeEvaluateLog:
		var resp Log
		return &resp, decoder.Decode(&resp)
	case codeNewEvaluatorResponse:
		var resp CreateEvaluatorResponse
		return &resp, decoder.Decode(&resp)
	case codeEvaluateRead:
		var resp ReadResource
		return &resp, decoder.Decode(&resp)
	case codeEvaluateReadModule:
		var resp ReadModule
		return &resp, decoder.Decode(&resp)
	case codeListResourcesRequest:
		var resp ListResources
		return &resp, decoder.Decode(&resp)
	case codeListModulesRequest:
		var resp ListModules
		return &resp, decoder.Decode(&resp)
	case codeInitializeModuleReaderRequest:
		var resp InitializeModuleReader
		return &resp, decoder.Decode(&resp)
	case codeInitializeResourceReaderRequest:
		var resp InitializeResourceReader
		return &resp, decoder.Decode(&resp)
	case codeCloseExternalProcess:
		var resp CloseExternalProcess
		return &resp, decoder.Decode(&resp)
	default:
		panic(fmt.Sprintf("Unknown code: %d", c))
	}
}
