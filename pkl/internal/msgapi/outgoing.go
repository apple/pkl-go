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
	"bytes"

	"github.com/vmihailenco/msgpack/v5"
)

type OutgoingMessage interface {
	ToMsgPack() ([]byte, error)
}

var (
	_ OutgoingMessage = (*CreateEvaluator)(nil)
	_ OutgoingMessage = (*CloseEvaluator)(nil)
	_ OutgoingMessage = (*Evaluate)(nil)
	_ OutgoingMessage = (*ReadResourceResponse)(nil)
	_ OutgoingMessage = (*ReadModuleResponse)(nil)
	_ OutgoingMessage = (*ListResourcesResponse)(nil)
	_ OutgoingMessage = (*ListModulesResponse)(nil)
)

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
	Env                     map[string]string          `msgpack:"env,omitempty"`
	Properties              map[string]string          `msgpack:"properties,omitempty"`
	Project                 *ProjectOrDependency       `msgpack:"project,omitempty"`
	Http                    *Http                      `msgpack:"http,omitempty"`
	ExternalModuleReaders   map[string]*ExternalReader `msgpack:"externalModuleReaders,omitempty"`
	ExternalResourceReaders map[string]*ExternalReader `msgpack:"externalResourceReaders,omitempty"`
	OutputFormat            string                     `msgpack:"outputFormat,omitempty"`
	RootDir                 string                     `msgpack:"rootDir,omitempty"`
	CacheDir                string                     `msgpack:"cacheDir,omitempty"`
	TraceMode               string                     `msgpack:"traceMode,omitempty"`
	ResourceReaders         []*ResourceReader          `msgpack:"clientResourceReaders,omitempty"`
	ModuleReaders           []*ModuleReader            `msgpack:"clientModuleReaders,omitempty"`
	ExternalReaderCommands  [][]string                 `msgpack:"externalReaderCommands,omitempty"`
	ModulePaths             []string                   `msgpack:"modulePaths,omitempty"`
	AllowedModules          []string                   `msgpack:"allowedModules,omitempty"`
	AllowedResources        []string                   `msgpack:"allowedResources,omitempty"`
	RequestId               int64                      `msgpack:"requestId"`
	// Intentionally not used right now. Go has `context.WithTimeout` which is a more canonical way to handle timeouts.
	TimeoutSeconds int64 `msgpack:"timeoutSeconds,omitempty"`
}

type ProjectOrDependency struct {
	Checksums      *Checksums                      `msgpack:"checksums,omitempty"`
	Dependencies   map[string]*ProjectOrDependency `msgpack:"dependencies"`
	PackageUri     string                          `msgpack:"packageUri,omitempty"`
	Type           string                          `msgpack:"type"`
	ProjectFileUri string                          `msgpack:"projectFileUri,omitempty"`
}

type Http struct {
	Proxy          *Proxy            `msgpack:"proxy,omitempty"`
	Rewrites       map[string]string `msgpack:"rewrites,omitempty"`
	CaCertificates []byte            `msgpack:"caCertificates,omitempty"`
}

type Proxy struct {
	Address string   `msgpack:"address,omitempty"`
	NoProxy []string `msgpack:"noProxy,omitempty"`
}

type ExternalReader struct {
	Executable string   `msgpack:"executable"`
	Arguments  []string `msgpack:"arguments,omitempty"`
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
	ModuleUri   string `msgpack:"moduleUri"`
	ModuleText  string `msgpack:"moduleText,omitempty"`
	Expr        string `msgpack:"expr,omitempty"`
	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
}

func (msg *Evaluate) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeEvaluate)
}

type ReadResourceResponse struct {
	Error       string `msgpack:"error,omitempty"`
	Contents    []byte `msgpack:"contents,omitempty"`
	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
}

func (msg *ReadResourceResponse) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeEvaluateReadResponse)
}

type ReadModuleResponse struct {
	Contents    string `msgpack:"contents,omitempty"`
	Error       string `msgpack:"error,omitempty"`
	RequestId   int64  `msgpack:"requestId"`
	EvaluatorId int64  `msgpack:"evaluatorId"`
}

func (msg *ReadModuleResponse) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeEvaluateReadModuleResponse)
}

type ListResourcesResponse struct {
	Error        string         `msgpack:"error,omitempty"`
	PathElements []*PathElement `msgpack:"pathElements,omitempty"`
	RequestId    int64          `msgpack:"requestId"`
	EvaluatorId  int64          `msgpack:"evaluatorId"`
}

func (msg ListResourcesResponse) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeListResourcesResponse)
}

type ListModulesResponse struct {
	Error        string         `msgpack:"error,omitempty"`
	PathElements []*PathElement `msgpack:"pathElements,omitempty"`
	RequestId    int64          `msgpack:"requestId"`
	EvaluatorId  int64          `msgpack:"evaluatorId"`
}

func (msg ListModulesResponse) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeListModulesResponse)
}

type PathElement struct {
	Name        string `msgpack:"name"`
	IsDirectory bool   `msgpack:"isDirectory"`
}

type InitializeModuleReaderResponse struct {
	Spec      *ModuleReader `msgpack:"spec,omitempty"`
	RequestId int64         `msgpack:"requestId"`
}

func (msg InitializeModuleReaderResponse) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeInitializeModuleReaderResponse)
}

type InitializeResourceReaderResponse struct {
	Spec      *ResourceReader `msgpack:"spec,omitempty"`
	RequestId int64           `msgpack:"requestId"`
}

func (msg InitializeResourceReaderResponse) ToMsgPack() ([]byte, error) {
	return packMessage(msg, codeInitializeResourceReaderResponse)
}
