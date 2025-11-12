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

package pkg

import (
	"context"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/go-cmp/cmp"

	"github.com/apple/pkl-go/cmd/pkl-gen-go/generatorsettings"
	"github.com/apple/pkl-go/pkl"
	"golang.org/x/mod/modfile"
)

//go:embed template.gopkl
var templateSource string

var (
	uriRegex = regexp.MustCompile("^[a-z]+:")
	tmpl     = template.Must(template.New("pkl").Parse(templateSource))
)

type TemplateValues struct {
	PklModulePath string
	generatorsettings.GeneratorSettings
}

func determineBasePath(v *generatorsettings.GeneratorSettings) error {
	if v.BasePath != "" {
		return nil
	}
	goModPath, _ := filepath.Abs("go.mod")
	goMod, err := os.ReadFile(goModPath)
	if err != nil {
		return nil
	}
	mod, err := modfile.Parse("go.mod", goMod, nil)
	if err != nil {
		return err
	}
	moduleName := mod.Module.Mod.Path
	fmt.Printf("Determined base path to be \033[33m%s\033[0m from go.mod\n", moduleName)
	v.BasePath = moduleName
	return nil
}

func doFormat(src string) ([]byte, string, error) {
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return nil, "", fmt.Errorf("error formatting Go source: %w", err)
	}
	strFormatted := string(formatted)
	if src == strFormatted {
		return formatted, "", nil
	}
	return formatted, cmp.Diff(src, strFormatted), nil
}

func generateDryRun(evaluator pkl.Evaluator, tmpFile *os.File, outputPath string, settings generatorsettings.GeneratorSettings) error {
	var filenames []string
	err := evaluator.EvaluateExpression(context.Background(), pkl.FileSource(tmpFile.Name()), "output.files.toMap().keys.toList()", &filenames)
	if err != nil {
		return err
	}
	log("Dry run; printing filenames but not writing files to disk\n")
	for _, filename := range filenames {
		if settings.BasePath != "" {
			if !strings.HasPrefix(filename, settings.BasePath) {
				continue
			}
			filename = strings.TrimPrefix(filename, settings.BasePath)
		}
		out := filepath.Join(outputPath, filename)
		fmt.Println(out)
	}
	return nil
}

func log(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func GenerateGo(
	evaluator pkl.Evaluator,
	pklModulePath string,
	settings generatorsettings.GeneratorSettings,
	silent bool,
	outputPath string,
) error {
	log("Generating Go sources for module \033[36m%s\033[0m\n", pklModulePath)
	var err error
	if !uriRegex.MatchString(pklModulePath) {
		pklModulePath, err = filepath.Abs(pklModulePath)
		if err != nil {
			return err
		}
	}
	if !strings.Contains(settings.GeneratorScriptPath, ":") {
		settings.GeneratorScriptPath, err = filepath.Abs(settings.GeneratorScriptPath)
		if err != nil {
			return err
		}
		log("Using custom generator script: \033[36m%s\033[0m\n", settings.GeneratorScriptPath)
	}
	if err = determineBasePath(&settings); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(os.TempDir(), "pkl-gen-go.*.pkl")
	if err != nil {
		return err
	}

	defer func() {
		if err := os.RemoveAll(tmpFile.Name()); err != nil {
			log("Failed to remove temporary file %s: %v\n", tmpFile.Name(), err)
		}
	}()

	templateValues := TemplateValues{
		GeneratorSettings: settings,
		PklModulePath:     pklModulePath,
	}
	if err = tmpl.Execute(tmpFile, templateValues); err != nil {
		return err
	}
	if settings.DryRun {
		return generateDryRun(evaluator, tmpFile, outputPath, settings)
	}
	files, err := evaluator.EvaluateOutputFiles(context.Background(), pkl.FileSource(tmpFile.Name()))
	if err != nil {
		return err
	}
	diffs := make(map[string]string)
	for filename, contents := range files {
		if settings.BasePath != "" {
			if !strings.HasPrefix(filename, settings.BasePath) {
				log("Skipping codegen for file \033[36m%s\033[0m because it does not exist in base path \033[36m%s\033[0m\n", filename, settings.BasePath)
				continue
			}
			filename = strings.TrimPrefix(filename, settings.BasePath)
		}

		formatted, diff, err := doFormat(contents)
		if err != nil {
			log("[warning] Attempted to format file %s but it produced an unexpected error. Error: %s\n", filename, err.Error())
			formatted = []byte(contents)
		}
		if diff != "" {
			diffs[filename] = diff
		}
		out := filepath.Join(outputPath, filename)
		if err = os.MkdirAll(filepath.Dir(out), 0o777); err != nil {
			return err
		}
		if err = os.WriteFile(out, formatted, 0o666); err != nil {
			return err
		}
		fmt.Println(out)
	}
	if len(diffs) > 0 && !silent {
		log("\n[notice] Some generated code needed to be formatted by the Go formatter." +
			" This is a bug on the Pkl side. To help us out, please file a GitHub issue for this, and include the Pkl source code input.\n" +
			"This message can be suppressed using the --suppress-format-warning flag.\n\n" +
			"Diffs:\n")
		for filename, diff := range diffs {
			fmt.Printf("\n")
			fmt.Printf("In file \033[36m%s\033[0m:\n", filename)
			fmt.Println(diff)
		}
	}
	return nil
}
