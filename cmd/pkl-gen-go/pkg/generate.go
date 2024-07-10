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

package pkg

import (
	"context"
	_ "embed"
	"fmt"
	"go/format"
	"log/slog"
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

var uriRegex = regexp.MustCompile("^[a-z]+:")

type TemplateValues struct {
	*generatorsettings.GeneratorSettings
	PklModulePath string
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
	var diffs string
	if src != strFormatted {
		diffs = cmp.Diff(src, strFormatted)
	}
	return formatted, diffs, nil
}

func generateDryRun(evaluator pkl.Evaluator, tmpFile *os.File, outputPath string, settings *generatorsettings.GeneratorSettings) error {
	var filenames []string
	err := evaluator.EvaluateExpression(context.Background(), pkl.FileSource(tmpFile.Name()), "output.files.toMap().keys.toList()", &filenames)
	if err != nil {
		return err
	}
	logWarning("Dry run; printing filenames but not writing files to disk\n")
	for _, filename := range filenames {
		if settings.BasePath != "" {
			if strings.HasPrefix(filename, settings.BasePath) {
				filename = strings.TrimPrefix(filename, settings.BasePath)
			} else {
				continue
			}
		}
		out := filepath.Join(outputPath, filename)
		fmt.Println(out)
	}
	return nil
}

func logInfo(format string, a ...any) {
	slog.Info(fmt.Sprintf(format, a...))
}

func logWarning(format string, a ...any) {
	slog.Warn(fmt.Sprintf(format, a...))
}

func GenerateGo(
	evaluator pkl.Evaluator,
	pklModulePath string,
	settings *generatorsettings.GeneratorSettings,
	silent bool,
	outputPath string,
) error {
	logInfo("Generating Go sources for module \033[36m%s\033[0m\n", pklModulePath)
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
		logInfo("Using custom generator script: \033[36m%s\033[0m\n", settings.GeneratorScriptPath)
	}
	if err = determineBasePath(settings); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(os.TempDir(), "pkl-gen-go.*.pkl")
	if err != nil {
		return err
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpFile.Name())
	tmpl, err := template.New("pkl").Parse(templateSource)
	if err != nil {
		return err
	}
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
			if strings.HasPrefix(filename, settings.BasePath) {
				filename = strings.TrimPrefix(filename, settings.BasePath)
			} else {
				logInfo("Skipping codegen for file \033[36m%s\033[0m because it does not exist in base path \033[36m%s\033[0m\n", filename, settings.BasePath)
				continue
			}
		}

		formatted, diff, err := doFormat(contents)
		if err != nil {
			logWarning("[warning] Attempted to format file %s but it produced an unexpected error. Error: %s\n", filename, err.Error())
			formatted = []byte(contents)
		}
		if len(diff) > 0 {
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
		logInfo("\n[notice] Some generated code needed to be formatted by the Go formatter." +
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
