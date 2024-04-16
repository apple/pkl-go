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

package pkl

import (
	"net/url"
	"os"
	"path"
)

// ModuleSource represents a source for Pkl evaluation.
type ModuleSource struct {
	// Uri is the URL of the resource.
	Uri *url.URL

	// Contents is the text contents of the resource, if any.
	//
	// If Contents is empty, it gets resolved by Pkl during evaluation time.
	// If the scheme of the Uri matches a ModuleReader, it will be used to resolve the module.
	Contents string
}

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

// TextSource builds a ModuleSource whose contents are the provided text.
func TextSource(text string) *ModuleSource {
	return &ModuleSource{
		// repl:text
		Uri: &url.URL{
			Scheme: "repl",
			Opaque: "text",
		},
		Contents: text,
	}
}

// UriSource builds a ModuleSource using the input uri.
//
// It panics if the uri is not valid.
func UriSource(uri string) *ModuleSource {
	parsedUri, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}
	return &ModuleSource{
		Uri: parsedUri,
	}
}
