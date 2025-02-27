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

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/apple/pkl-go/cmd/pkl-gen-go/generatorsettings"
	"github.com/apple/pkl-go/cmd/pkl-gen-go/pkg"
	"github.com/apple/pkl-go/pkl"
)

var lineNoRegex = regexp.MustCompile(`(?m)^(( ║ )*)(\d+) \|`)

// stripLineNumbers replaces line numbers with the same amount of the letter "x"
// (to preserve formatting).
func stripLineNumbers(output string) string {
	return lineNoRegex.ReplaceAllStringFunc(output, func(s string) string {
		return strings.Repeat("x", len(s)-2) + " |"
	})
}

func makeGoCode(evaluator pkl.Evaluator, snippetsDir string) {
	outputDir := filepath.Join(snippetsDir, "output")
	if err := os.RemoveAll(outputDir); err != nil {
		panic(err)
	}
	inputDir := filepath.Join(snippetsDir, "input")
	files, err := os.ReadDir(inputDir)
	if err != nil {
		panic(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	codegenDir := filepath.Join(snippetsDir, "..")
	settings, err := generatorsettings.LoadFromPath(context.Background(), "codegen/snippet-tests/generator-settings.pkl")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		err := pkg.GenerateGo(evaluator, filepath.Join(inputDir, file.Name()), settings, false, cwd)
		if strings.Contains(file.Name(), "err.pkl") {
			if err == nil {
				fmt.Printf("ERROR: Expected %s to error, but it did not\n", file.Name())
				os.Exit(1)
			}
			basename := strings.TrimSuffix(filepath.Base(file.Name()), ".pkl")
			errContents := strings.ReplaceAll(err.Error(), codegenDir, "<codegen_dir>")
			errContents = stripLineNumbers(errContents)
			if err = os.WriteFile(filepath.Join(outputDir, basename), []byte(errContents), 0o666); err != nil {
				panic(err)
			}
		} else if err != nil {
			panic(err)
		}
	}
}

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("can't find caller")
	}
	snippetsDir := filepath.Join(filepath.Dir(filename), "../../../codegen/snippet-tests")
	evaluator, err := pkl.NewEvaluator(context.Background(), pkl.PreconfiguredOptions)
	if err != nil {
		panic(err)
	}
	makeGoCode(evaluator, snippetsDir)
}
