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

const (
	codeNewEvaluator               int = 0x20
	codeNewEvaluatorResponse       int = 0x21
	codeCloseEvaluator             int = 0x22
	codeEvaluate                   int = 0x23
	codeEvaluateResponse           int = 0x24
	codeEvaluateLog                int = 0x25
	codeEvaluateRead               int = 0x26
	codeEvaluateReadResponse       int = 0x27
	codeEvaluateReadModule         int = 0x28
	codeEvaluateReadModuleResponse int = 0x29
	codeListResourcesRequest       int = 0x2a
	codeListResourcesResponse      int = 0x2b
	codeListModulesRequest         int = 0x2c
	codeListModulesResponse        int = 0x2d
)
