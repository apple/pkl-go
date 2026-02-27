//===----------------------------------------------------------------------===//
// Copyright © 2026 Apple Inc. and the Pkl project authors. All rights reserved.
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

package main

import (
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/apple/pkl-go/pkl"
)

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("can't find caller")
	}

	client, err := pkl.NewExternalReaderClient(pkl.WithExternalClientModuleReader(fixtureReader{
		baseDir: filepath.Join(filepath.Dir(filename), "../../.."),
	}))
	if err != nil {
		log.Fatalln(err)
	}
	if err := client.Run(); err != nil {
		log.Fatalln(err)
	}
}

type fixtureReader struct {
	baseDir string
}

func (f fixtureReader) Scheme() string {
	return "pklgo"
}

func (f fixtureReader) IsGlobbable() bool {
	return true
}

func (f fixtureReader) HasHierarchicalUris() bool {
	return true
}

func (f fixtureReader) ListElements(url url.URL) ([]pkl.PathElement, error) {
	path := strings.Trim(url.Path, "/")
	if path == "" {
		path = "."
	}
	entries, err := os.ReadDir(filepath.Join(f.baseDir, path))
	if err != nil {
		return nil, err
	}
	ret := make([]pkl.PathElement, 0, len(entries))
	for _, entry := range entries {
		// copy Pkl's built-in `file` ModuleKey and don't follow symlinks.
		if entry.Type()&fs.ModeSymlink != 0 {
			continue
		}
		ret = append(ret, pkl.NewPathElement(entry.Name(), entry.IsDir()))
	}
	return ret, nil
}

func (f fixtureReader) IsLocal() bool {
	return true
}

func (f fixtureReader) Read(url url.URL) (string, error) {
	contents, err := os.ReadFile(filepath.Join(f.baseDir, strings.TrimPrefix(url.Path, "/")))
	return string(contents), err
}
