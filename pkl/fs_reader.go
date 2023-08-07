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
	"io/fs"
	"net/url"
	"strings"
)

type fsReader struct {
	fs     fs.FS
	scheme string
}

func (f *fsReader) Scheme() string {
	return f.scheme
}

func (f *fsReader) IsGlobbable() bool {
	return true
}

func (f *fsReader) HasHierarchicalUris() bool {
	return true
}

func (f *fsReader) ListElements(url url.URL) ([]PathElement, error) {
	path := strings.TrimSuffix(strings.TrimPrefix(url.Path, "/"), "/")
	if path == "" {
		path = "."
	}
	entries, err := fs.ReadDir(f.fs, path)
	if err != nil {
		return nil, err
	}
	var ret []PathElement
	for _, entry := range entries {
		// copy Pkl's built-in `file` ModuleKey and don't follow symlinks.
		if entry.Type()&fs.ModeSymlink != 0 {
			continue
		}
		ret = append(ret, NewPathElement(entry.Name(), entry.IsDir()))
	}
	return ret, nil
}

var _ Reader = (*fsReader)(nil)

type fsModuleReader struct {
	*fsReader
}

func (f fsModuleReader) IsLocal() bool {
	return true
}

func (f fsModuleReader) Read(url url.URL) (string, error) {
	contents, err := fs.ReadFile(f.fs, strings.TrimPrefix(url.Path, "/"))
	return string(contents), err
}

var _ ModuleReader = (*fsModuleReader)(nil)

type fsResourceReader struct {
	*fsReader
}

func (f fsResourceReader) Read(url url.URL) ([]byte, error) {
	return fs.ReadFile(f.fs, strings.TrimPrefix(url.Path, "/"))
}

var _ ResourceReader = (*fsResourceReader)(nil)
