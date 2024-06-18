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

//go:build unix
package pkl

import (
	"net/url"
	"os"
	"path"
)

// FileSource builds a ModuleSource, treating its arguments as paths on the file system.
//
// If the provided path is not an absolute path, it will be resolved against the current working
// directory.
//
// If multiple path arguments are provided, they are joined as multiple elements of the path.
//
// It panics if the current working directory cannot be resolved.
func FileSource(pathElems ...string) *ModuleSource {
	src := path.Join(pathElems...)
	if !path.IsAbs(src) {
		p, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		src = path.Join(p, src)
	}
	return &ModuleSource{
		Uri: &url.URL{
			Scheme: "file",
			Path:   src,
		},
	}
}
