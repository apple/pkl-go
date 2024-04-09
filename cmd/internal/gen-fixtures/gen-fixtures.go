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
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"runtime"

	"github.com/apple/pkl-go/cmd/pkl-gen-go/generatorsettings"
	"github.com/apple/pkl-go/cmd/pkl-gen-go/pkg"
	"github.com/apple/pkl-go/pkl"
)

type testFs struct{}

var _ fs.FS = (*testFs)(nil)

func (t testFs) Open(name string) (fs.File, error) {
	_, fileName, _, _ := runtime.Caller(0)
	return os.Open(path.Join(fileName, "../../../../", name))
}

func evaluateCollections(evaluator pkl.Evaluator, fixturesDir string) {
	for _, expr := range []string{"res1", "res2", "res9"} {
		outBytes, err := evaluator.EvaluateExpressionRaw(context.Background(), pkl.FileSource(fixturesDir, "collections.pkl"), expr)
		if err != nil {
			panic(err)
		}
		outPath := path.Join(fixturesDir, "msgpack", fmt.Sprintf("collections.%s.msgpack", expr))
		if err = os.WriteFile(outPath, outBytes, 0o666); err != nil {
			panic(err)
		}
		fmt.Printf("Wrote file %s\n", outPath)
	}
}

func makeMsgpack(evaluator pkl.Evaluator, fixturesDir string, files []os.DirEntry) {
	msgpackDir := path.Join(fixturesDir, "msgpack")
	if err := os.RemoveAll(msgpackDir); err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		outBytes, err := evaluator.EvaluateExpressionRaw(context.Background(), pkl.UriSource("pklgo:/pkl/test_fixtures/"+file.Name()), "")
		if err != nil {
			panic(err)
		}
		outPath := path.Join(fixturesDir, "msgpack", file.Name()+".msgpack")
		if err = os.MkdirAll(path.Dir(outPath), 0o750); err != nil {
			panic(err)
		}
		if err = os.WriteFile(outPath, outBytes, 0o600); err != nil {
			panic(err)
		}
		fmt.Printf("Wrote file %s\n", outPath)
	}
	evaluateCollections(evaluator, fixturesDir)
}

func makeGoCode(evaluator pkl.Evaluator, fixturesDir string, files []os.DirEntry) {
	genDir := path.Join(fixturesDir, "gen")
	if err := os.RemoveAll(genDir); err != nil {
		panic(err)
	}
	settings, err := generatorsettings.LoadFromPath(context.Background(), "codegen/snippet-tests/generator-settings.pkl")
	if err != nil {
		panic(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if err := pkg.GenerateGo(evaluator, path.Join(fixturesDir, file.Name()), settings, false, cwd); err != nil {
			panic(err)
		}
	}
}

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("can't find caller")
	}
	fixturesDir := path.Join(path.Dir(filename), "../../../pkl/test_fixtures")
	files, err := os.ReadDir(fixturesDir)
	if err != nil {
		panic(err)
	}
	manager := pkl.NewEvaluatorManager()
	evaluator, err := manager.NewEvaluator(
		context.Background(),
		pkl.PreconfiguredOptions,
		// Coerce the module URI for fixtures to have scheme `pklgo` and resolve off the project
		// root. For example: `pklgo:/pkl/test_fixtures/any.pkl`
		// We do this so that these URIs don't change from machine to machine.
		pkl.WithFs(testFs{}, "pklgo"),
	)
	if err != nil {
		panic(err)
	}
	makeMsgpack(evaluator, fixturesDir, files)
	makeGoCode(evaluator, fixturesDir, files)
}
