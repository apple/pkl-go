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

var _ IncomingMessage = (*CreateEvaluatorResponse)(nil)
var _ IncomingMessage = (*EvaluateResponse)(nil)
var _ IncomingMessage = (*ReadResource)(nil)
var _ IncomingMessage = (*ReadModule)(nil)
var _ IncomingMessage = (*Log)(nil)
var _ IncomingMessage = (*ListResources)(nil)
var _ IncomingMessage = (*ListModules)(nil)

type CreateEvaluatorResponse struct {
	incomingMessageImpl

	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
	Error       string `msgpack:"error"`
}

type EvaluateResponse struct {
	incomingMessageImpl

	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
	Result      []byte `msgpack:"result"`
	Error       string `msgpack:"error"`
}

type ReadResource struct {
	incomingMessageImpl

	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
	Uri         string `msgpack:"uri"`
}

type ReadModule struct {
	incomingMessageImpl

	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
	Uri         string `msgpack:"uri"`
}

type Log struct {
	incomingMessageImpl

	EvaluatorId int64  `msgpack:"evaluatorId"`
	Level       int    `msgpack:"level"`
	Message     string `msgpack:"message"`
	FrameUri    string `msgpack:"frameUri"`
}

type ListResources struct {
	incomingMessageImpl

	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
	Uri         string `msgpack:"uri"`
}

type ListModules struct {
	incomingMessageImpl

	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
	Uri         string `msgpack:"uri"`
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
		err = decoder.Decode(&resp)
		return &resp, err
	case codeEvaluateLog:
		var resp Log
		err = decoder.Decode(&resp)
		return &resp, err
	case codeNewEvaluatorResponse:
		var resp CreateEvaluatorResponse
		err = decoder.Decode(&resp)
		return &resp, err
	case codeEvaluateRead:
		var resp ReadResource
		err = decoder.Decode(&resp)
		return &resp, err
	case codeEvaluateReadModule:
		var resp ReadModule
		err = decoder.Decode(&resp)
		return &resp, err
	case codeListResourcesRequest:
		var resp ListResources
		err = decoder.Decode(&resp)
		return &resp, err
	case codeListModulesRequest:
		var resp ListModules
		err = decoder.Decode(&resp)
		return &resp, err
	default:
		panic(fmt.Sprintf("Unknown code: %d", int(c)))
	}
}
