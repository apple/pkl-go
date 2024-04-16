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

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"runtime"
	"strings"
)

const licenseHeader = `// ===----------------------------------------------------------------------===//
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
`

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("can't find caller")
	}
	projectRoot := path.Join(path.Dir(filename), "../../../")
	projectRootFs := os.DirFS(projectRoot)
	var files []string
	err := fs.WalkDir(projectRootFs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".pkl") || strings.HasSuffix(d.Name(), ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		absolutePath := path.Join(projectRoot, file)
		contents, err := os.ReadFile(absolutePath)
		if err != nil {
			panic(err)
		}
		contentsStr := string(contents)
		if strings.HasPrefix(contentsStr, licenseHeader) {
			continue
		}
		if strings.HasPrefix(contentsStr, "// Code generated from") {
			continue
		}
		var newContents string
		if strings.HasSuffix(file, ".go") {
			newContents = licenseHeader + "\n" + contentsStr
		} else {
			newContents = licenseHeader + contentsStr
		}
		if err = os.WriteFile(absolutePath, []byte(newContents), 0o644); err != nil {
			panic(err)
		}
		fmt.Println("Wrote license header to " + file)
	}
}
