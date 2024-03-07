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
	"bytes"

	"github.com/vmihailenco/msgpack/v5"
)

type OutgoingMessage interface {
	ToMsgPack() ([]byte, error)
}

var _ OutgoingMessage = (*CreateEvaluator)(nil)
var _ OutgoingMessage = (*CloseEvaluator)(nil)
var _ OutgoingMessage = (*Evaluate)(nil)
var _ OutgoingMessage = (*ReadResourceResponse)(nil)
var _ OutgoingMessage = (*ReadModuleResponse)(nil)
var _ OutgoingMessage = (*ListResourcesResponse)(nil)
var _ OutgoingMessage = (*ListModulesResponse)(nil)

func packMessage(msg OutgoingMessage, code int) ([]byte, error) {
	enc := msgpack.NewEncoder(nil)
	var buf bytes.Buffer
	enc.Reset(&buf)
	if err := enc.EncodeArrayLen(2); err != nil {
		return nil, err
	}
	if err := enc.EncodeInt(int64(code)); err != nil {
		return nil, err
	}
	if err := enc.Encode(msg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type ResourceReader struct {
	Scheme              string `msgpack:"scheme"`
	HasHierarchicalUris bool   `msgpack:"hasHierarchicalUris"`
	IsGlobbable         bool   `msgpack:"isGlobbable"`
}

type ModuleReader struct {
	Scheme              string `msgpack:"scheme"`
	HasHierarchicalUris bool   `msgpack:"hasHierarchicalUris"`
	IsGlobbable         bool   `msgpack:"isGlobbable"`
	IsLocal             bool   `msgpack:"isLocal"`
}

type CreateEvaluator struct {
	RequestId        int64                `msgpack:"requestId"`
	ResourceReaders  []*ResourceReader    `msgpack:"clientResourceReaders,omitempty"`
	ModuleReaders    []*ModuleReader      `msgpack:"clientModuleReaders,omitempty"`
	ModulePaths      []string             `msgpack:"modulePaths,omitempty"`
	Env              map[string]string    `msgpack:"env,omitempty"`
	Properties       map[string]string    `msgpack:"properties,omitempty"`
	OutputFormat     string               `msgpack:"outputFormat,omitempty"`
	AllowedModules   []string             `msgpack:"allowedModules,omitempty"`
	AllowedResources []string             `msgpack:"allowedResources,omitempty"`
	RootDir          string               `msgpack:"rootDir,omitempty"`
	CacheDir         string               `msgpack:"cacheDir,omitempty"`
	Project          *ProjectOrDependency `msgpack:"project,omitempty"`
	// Intentionally not used right now. Go has `context.WithTimeout` which is a more canonical way to handle timeouts.
	TimeoutSeconds int64 `msgpack:"timeoutSeconds,omitempty"`
}

type ProjectOrDependency struct {
	PackageUri     string                          `msgpack:"packageUri,omitempty"`
	Type           string                          `msgpack:"type"`
	ProjectFileUri string                          `msgpack:"projectFileUri,omitempty"`
	Checksums      *Checksums                      `msgpack:"checksums,omitempty"`
	Dependencies   map[string]*ProjectOrDependency `msgpack:"dependencies"`
}

type Checksums struct {
	Sha256 string `msgpack:"checksums"`
}

func (msg *CreateEvaluator) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeNewEvaluator)
}

type CloseEvaluator struct {
	EvaluatorId int64 `msgpack:"evaluatorId,omitempty"`
}

func (msg *CloseEvaluator) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeCloseEvaluator)
}

type Evaluate struct {
	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
	ModuleUri   string `msgpack:"moduleUri"`
	ModuleText  string `msgpack:"moduleText,omitempty"`
	Expr        string `msgpack:"expr,omitempty"`
}

func (msg *Evaluate) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeEvaluate)
}

type ReadResourceResponse struct {
	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
	Contents    []byte `msgpack:"contents,omitempty"`
	Error       string `msgpack:"error,omitempty"`
}

func (msg *ReadResourceResponse) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeEvaluateReadResponse)
}

type ReadModuleResponse struct {
	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
	Contents    string `msgpack:"contents,omitempty"`
	Error       string `msgpack:"error,omitempty"`
}

func (msg *ReadModuleResponse) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeEvaluateReadModuleResponse)
}

type ListResourcesResponse struct {
	RequestId    int64          `msgpack:"requestId"`
	EvaluatorId  int64          `msgpack:"evaluatorId"`
	PathElements []*PathElement `msgpack:"pathElements,omitempty"`
	Error        string         `msgpack:"error,omitempty"`
}

func (msg ListResourcesResponse) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeListResourcesResponse)
}

type ListModulesResponse struct {
	RequestId    int64          `msgpack:"requestId"`
	EvaluatorId  int64          `msgpack:"evaluatorId"`
	PathElements []*PathElement `msgpack:"pathElements,omitempty"`
	Error        string         `msgpack:"error,omitempty"`
}

func (msg ListModulesResponse) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeListModulesResponse)
}

type PathElement struct {
	Name        string `msgpack:"name"`
	IsDirectory bool   `msgpack:"isDirectory"`
}
