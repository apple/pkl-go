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
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/apple/pkl-go/cmd/pkl-gen-go/generatorsettings"
	"github.com/apple/pkl-go/cmd/pkl-gen-go/pkg"
	"github.com/apple/pkl-go/pkl"
)

var command = cobra.Command{
	Use:   "pkl-gen-go [flags] <module>",
	Short: "Generates Go bindings for a Pkl module",
	Long: `Generates Go bindings for a Pkl module.

PACKAGE MAPPINGS
	To generate Go, all Pkl modules must have a known Go package name. The package name
	may come from one of three sources:

	  1. The @go.Package annotation on a module
	  2. A generator settings Pkl file
	  3. A --mapping argument

GENERATOR SETTINGS FILE
	Code generation may be configured using a settings file. By default, pkl-gen-go will look 
	for file called "generator-settings.pkl" in the current working directory, and the path can
	be configured using the --generator-settings flag.

	The generator settings file should amend module 
	package://pkg.pkl-lang.org/pkl-go/pkl.golang@<VERSION>#/GeneratorSettings.pkl

CONFIGURING OUTPUT PATH
	By default, the full path of each module is written as a relative path to the current working
	directory. This behavior changes by setting a base path either as a CLI flag, or in the
	generator settings file.

	When using a base path, any package that does not belong to the path will be skipped from
	code generation.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if printVersion {
			fmt.Println(Version)
			return nil
		}
		evaluator, err := newEvaluator()
		if err != nil {
			return fmt.Errorf("failed to create evaluator: %w", err)
		}
		if outputPath == "" {
			outputPath, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		if err = pkg.GenerateGo(evaluator, args[0], settings, suppressWarnings, outputPath); err != nil {
			_, _ = fmt.Fprint(os.Stderr, err.Error())
			os.Exit(1)
		}
		return nil
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if printVersion {
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
}

func determineProjectDir() string {
	projectDir := ""
	if settings.ProjectDir != nil {
		if filepath.IsAbs(*settings.ProjectDir) {
			projectDir = *settings.ProjectDir
		} else {
			// If settings are not overriden - it was taken from CWD
			if settings.GeneratorSettingsPath == nil {
				projectDir = path.Join(settings.Cwd, *settings.ProjectDir)
			} else { // If settings are somewhere else, we shall build project relative path from settings file
				settingPath := *settings.GeneratorSettingsPath
				if filepath.IsAbs(settingPath) { // we shall not use CWD for absolute path
					projectDir = path.Join(path.Dir(settingPath), *settings.ProjectDir)
				} else { // for relative settings path we shall adjust project dir to settings file location, rather than CWD
					projectDir = path.Join(settings.Cwd, path.Dir(settingPath), *settings.ProjectDir)
				}
			}
		}
	}

	return projectDir
}

func newEvaluator() (pkl.Evaluator, error) {
	projectDirFound := findProjectDir(determineProjectDir())
	if projectDirFound == "" {
		return pkl.NewEvaluator(context.Background(), evaluatorOptions)
	}
	return pkl.NewProjectEvaluator(context.Background(), projectDirFound, evaluatorOptions)
}

func evaluatorOptions(opts *pkl.EvaluatorOptions) {
	pkl.MaybePreconfiguredOptions(opts)
	opts.Logger = pkl.StderrLogger
	if len(settings.AllowedModules) > 0 {
		opts.AllowedModules = settings.AllowedModules
	}
	if len(settings.AllowedResources) > 0 {
		opts.AllowedResources = settings.AllowedResources
	}
}

var (
	settings         *generatorsettings.GeneratorSettings
	suppressWarnings bool
	outputPath       string
	printVersion     bool
)

// The version of pkl-gen-go.
//
// This gets replaced by ldflags when built through CI,
// or by init when installed via go install.
var Version = "development"

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || Version != "development" {
		return
	}
	Version = strings.TrimPrefix(info.Main.Version, "v")
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	} else if err != nil {
		panic(err)
	}
	return true
}

//goland:noinspection GoBoolExpressions
func generatorSettingsSource() *pkl.ModuleSource {
	// (devel) may be returned from init()
	if Version == "development" || Version == "(devel)" {
		_, filename, _, ok := runtime.Caller(1)
		if !ok {
			panic("Failed to get path to pkl-gen-go.go")
		}
		dirPath := filepath.Dir(filename)
		return pkl.FileSource(dirPath, "../../codegen/src/GeneratorSettings.pkl")
	}
	return pkl.UriSource(fmt.Sprintf("package://pkg.pkl-lang.org/pkl-go/pkl.golang@%s#/GeneratorSettings.pkl", Version))
}

// mimick logic for finding project dir in the pkl CLI.
func doFindProjectDir(dir string) string {
	if fileExists(filepath.Join(dir, "PklProject")) {
		return dir
	}
	parent := filepath.Dir(dir)
	if parent == dir {
		return ""
	}
	return doFindProjectDir(parent)
}

func findProjectDir(projectDirFlag string) string {
	if projectDirFlag != "" {
		return projectDirFlag
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return doFindProjectDir(cwd)
}

// Loads the settings for controlling codegen.
// Uses a Pkl evaluator that is separate from what's used for actually running codegen.
func loadGeneratorSettings(generatorSettingsPath string, projectDirFlag string) (*generatorsettings.GeneratorSettings, error) {
	projectDir := findProjectDir(projectDirFlag)
	var evaluator pkl.Evaluator
	var err error
	if projectDir != "" {
		evaluator, err = pkl.NewProjectEvaluator(context.Background(), projectDir, pkl.PreconfiguredOptions)
	} else {
		evaluator, err = pkl.NewEvaluator(context.Background(), pkl.PreconfiguredOptions)
	}
	if err != nil {
		panic(err)
	}
	var source *pkl.ModuleSource
	if generatorSettingsPath != "" {
		source = pkl.FileSource(generatorSettingsPath)
	} else if fileExists("generator-settings.pkl") {
		source = pkl.FileSource("generator-settings.pkl")
	} else {
		source = generatorSettingsSource()
	}
	return generatorsettings.Load(context.Background(), evaluator, source)
}

func init() {
	flags := command.Flags()
	var generatorSettingsPath string
	var generateScript string
	var mappings map[string]string
	var basePath string
	var allowedModules []string
	var allowedResources []string
	var dryRun bool
	var projectDir string
	flags.StringVar(&generatorSettingsPath, "generator-settings", "", "The path to a generator settings file. Default: current working directory.")
	flags.StringVar(&generateScript, "generate-script", "", "The Generate.pkl script to use")
	flags.StringToStringVar(&mappings, "mapping", nil, "The mapping of a Pkl module name to a Go package name")
	flags.StringVar(&basePath, "base-path", "", "The base path used to determine relative output")
	flags.StringVar(&outputPath, "output-path", "", "The output directory to write generated sources into")
	flags.BoolVar(&suppressWarnings, "suppress-format-warning", false, "Suppress warnings around formatting issues")
	flags.StringSliceVar(&allowedModules, "allowed-modules", nil, "URI patterns that determine which modules can be loaded and evaluated")
	flags.StringSliceVar(&allowedResources, "allowed-resources", nil, "URI patterns that determine which resources can be loaded and evaluated")
	flags.StringVar(&projectDir, "project-dir", "", "The project directory to load dependency and evaluator settings from. For relative paths workdir is determined based on --generator-settings.")
	flags.BoolVar(&dryRun, "dry-run", false, "Print out the names of the files that will be generated, but don't write any files")
	flags.BoolVar(&printVersion, "version", false, "Print the version and exit")
	var err error
	if err = flags.Parse(os.Args); err != nil && !errors.Is(err, pflag.ErrHelp) {
		panic(err)
	}
	settings, err = loadGeneratorSettings(generatorSettingsPath, projectDir)
	if err != nil {
		panic(err)
	}
	if generateScript != "" {
		settings.GeneratorScriptPath = generateScript
	}
	if len(mappings) != 0 {
		settings.PackageMappings = mappings
	}
	if basePath != "" {
		settings.BasePath = basePath
	}
	if len(allowedModules) > 0 {
		settings.AllowedModules = allowedModules
	}
	if len(allowedResources) > 0 {
		settings.AllowedResources = allowedResources
	}
	if projectDir != "" {
		settings.ProjectDir = &projectDir
	}
	settings.DryRun = dryRun

	if generatorSettingsPath != "" {
		settings.GeneratorSettingsPath = &generatorSettingsPath
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	settings.Cwd = cwd
}

func main() {
	if err := command.Execute(); err != nil {
		panic(err)
	}
}
