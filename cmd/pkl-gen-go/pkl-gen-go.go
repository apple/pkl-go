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
	"errors"
	"fmt"
	"github.com/apple/pkl-go/cmd/pkl-gen-go/generatorsettings"
	"github.com/apple/pkl-go/cmd/pkl-gen-go/pkg"
	"github.com/apple/pkl-go/pkl"
	"github.com/spf13/cobra"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
)

// Version is the version of pkl-gen-go that is built into the binary.
//
// This gets replaced by ldflags when built through CI, or by init when installed via go install.
var Version = "development"

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || Version != "development" {
		return
	}
	Version = strings.TrimPrefix(info.Main.Version, "v")
}

// stringErr is a string that satisfies the error interface.
//
// Doing this lets us define constant strings as errors.
type stringErr string

func (s stringErr) Error() string {
	return string(s)
}

const (
	// ErrMalformedArgs is returned when the user provides CLI arguments that are badly formed in some way.
	ErrMalformedArgs = stringErr("malformed arguments")

	// ErrRuntimeCallerFailure is returned when there is a problem using Go's runtime caller functionality.
	ErrRuntimeCallerFailure = stringErr("runtime caller failure")
)

// long command description
const longDescription = `Generates Go bindings for a Pkl module.

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
`

// root cobra command for pkl-gen-go
var command = cobra.Command{
	Use:     "pkl-gen-go [flags] <module>",
	Short:   "Generates Go bindings for a Pkl module",
	Long:    longDescription,
	PreRunE: commandPreRunE,
	RunE:    commandRunE,
}

// flag names
const (
	flagNamePrintVersion      = "version"
	flagNameBasePath          = "base-path"
	flagNameGeneratorSettings = "generator-settings"
	flagNameGenerateScript    = "generate-script"
	flagNameProjectDirname    = "project-dir"
	flagNameSuppressWarnings  = "suppress-format-warning"
	flagNameOutputPath        = "output-path"
	flagNamePackageMappings   = "mapping"
	flagNameAllowedModules    = "allowed-modules"
	flagNameAllowedResources  = "allowed-resources"
	flagNameDryRun            = "dry-run"
)

// initialize command flag set
func init() {
	// find the current working directory
	cwd, cwdErr := os.Getwd()
	handleFatalError(cwdErr)

	// define root command flags
	flagSet := command.Flags()
	flagSet.Bool(flagNamePrintVersion, false, "Print the version and exit")
	flagSet.String(flagNameBasePath, "", "The base path used to determine relative output")
	flagSet.String(flagNameGeneratorSettings, "", "path to a generator settings file")
	flagSet.String(flagNameGenerateScript, "", "The Generate.pkl script to use")
	flagSet.String(flagNameProjectDirname, "", "project directory from which dependency and evaluator settings are loaded")
	flagSet.Bool(flagNameSuppressWarnings, false, "Suppress warnings around formatting issues")
	flagSet.String(flagNameOutputPath, cwd, "The output directory to write generated sources into")
	flagSet.StringToString(flagNamePackageMappings, nil, "The mapping of a Pkl module name to a Go package name")
	flagSet.StringSlice(flagNameAllowedModules, nil, "URI patterns that determine which modules can be loaded and evaluated")
	flagSet.StringSlice(flagNameAllowedResources, nil, "URI patterns that determine which resources can be loaded and evaluated")
	flagSet.Bool(flagNameDryRun, false, "Print out the names of the files that will be generated, but don't write any files")
}

// context keys
type ckGeneratorSettings struct{}
type ckSuppressWarnings struct{}
type ckOutputPath struct{}

// command pre-run logic
func commandPreRunE(cmd *cobra.Command, args []string) error {
	flagSet := cmd.Flags()

	// resolve flag values
	printVersion := unwrapValue(flagSet.GetBool(flagNamePrintVersion))
	basePath := unwrapValue(flagSet.GetString(flagNameBasePath))
	generatorSettingsFilename := unwrapValue(flagSet.GetString(flagNameGeneratorSettings))
	generateScript := unwrapValue(flagSet.GetString(flagNameGenerateScript))
	projectDirname := unwrapValue(flagSet.GetString(flagNameProjectDirname))
	suppressWarnings := unwrapValue(flagSet.GetBool(flagNameSuppressWarnings))
	outputPath := unwrapValue(flagSet.GetString(flagNameOutputPath))
	packageMappings := unwrapValue(flagSet.GetStringToString(flagNamePackageMappings))
	allowedModules := unwrapValue(flagSet.GetStringSlice(flagNameAllowedModules))
	allowedResources := unwrapValue(flagSet.GetStringSlice(flagNameAllowedResources))
	dryRun := unwrapValue(flagSet.GetBool(flagNameDryRun))

	// if the user wants to print the version and exit, just do that now so that we don't need to waste resources
	// loading everything
	if printVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	// expect at exactly one argument
	if l := len(args); l != 1 {
		return fmt.Errorf("%w: must provide exactly one argument", ErrMalformedArgs)
	}

	// initialize generator settings
	settings, settingsErr := loadGeneratorSettings(generatorSettingsFilename, projectDirname)
	if settingsErr != nil {
		return settingsErr
	}

	// load generator settings that are set from flag values
	if projectDirname != "" {
		settings.ProjectDir = &projectDirname
	}
	if basePath != "" {
		settings.BasePath = basePath
	}
	if generateScript != "" {
		settings.GeneratorScriptPath = generateScript
	}
	if len(packageMappings) > 0 {
		settings.PackageMappings = packageMappings
	}
	if len(allowedModules) > 0 {
		settings.AllowedModules = allowedModules
	}
	if len(allowedResources) > 0 {
		settings.AllowedResources = allowedResources
	}
	settings.DryRun = dryRun

	// store context values
	ctx := cmd.Context()
	ctx = context.WithValue(ctx, ckGeneratorSettings{}, settings)
	ctx = context.WithValue(ctx, ckSuppressWarnings{}, suppressWarnings)
	ctx = context.WithValue(ctx, ckOutputPath{}, outputPath)
	cmd.SetContext(ctx)

	return nil
}

// command business logic
func commandRunE(cmd *cobra.Command, args []string) error {

	// pull context values out of the command context
	ctx := cmd.Context()
	generatorSettings := ctx.Value(ckGeneratorSettings{}).(*generatorsettings.GeneratorSettings)
	suppressWarnings := ctx.Value(ckSuppressWarnings{}).(bool)
	outputPath := ctx.Value(ckOutputPath{}).(string)

	// create the main evaluator
	evaluator, evalErr := newEvaluator(generatorSettings)
	if evalErr != nil {
		return evalErr
	}

	// generate Go code using the evaluator, and whatever other directly provided parameters are needed
	return pkg.GenerateGo(evaluator, args[0], generatorSettings, suppressWarnings, outputPath)
}

func fileExists(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// findProjectDir mimics logic for finding project dir in the pkl CLI.
func findProjectDir(dir string) (string, error) {
	exists, existsErr := fileExists(filepath.Join(dir, "PklProject"))
	if existsErr != nil {
		return "", existsErr
	} else if exists {
		return dir, nil
	}
	parent := filepath.Dir(dir)
	if parent == dir {
		return "", nil
	}
	return findProjectDir(parent)
}

// closure that produces a functional evaluator options modifier using the given generator settings
func evaluatorOptions(settings *generatorsettings.GeneratorSettings) func(opts *pkl.EvaluatorOptions) {
	return func(opts *pkl.EvaluatorOptions) {
		pkl.MaybePreconfiguredOptions(opts)
		opts.Logger = pkl.StderrLogger
		if len(settings.AllowedModules) > 0 {
			opts.AllowedModules = settings.AllowedModules
		}
		if len(settings.AllowedResources) > 0 {
			opts.AllowedResources = settings.AllowedResources
		}
	}
}

// newEvaluator creates the main Pkl evaluator
func newEvaluator(settings *generatorsettings.GeneratorSettings) (pkl.Evaluator, error) {
	var projectDir string

	// find the configured project directory
	if settings.ProjectDir != nil {
		if filepath.IsAbs(*settings.ProjectDir) {
			projectDir = *settings.ProjectDir
		} else {
			settingsUri, err := url.Parse(settings.Uri)
			if err != nil {
				return nil, fmt.Errorf("failed to parse settings.pkl URI: %w", err)
			}
			projectDir = path.Join(settingsUri.Path, "..", *settings.ProjectDir)
		}
	}

	// resolve the project directory
	resolvedProjectDir, dirErr := findProjectDir(projectDir)
	if dirErr != nil {
		return nil, dirErr
	}

	// create the appropriate evaluator
	if resolvedProjectDir == "" {
		return pkl.NewEvaluator(context.Background(), evaluatorOptions(settings))
	}
	return pkl.NewProjectEvaluator(context.Background(), resolvedProjectDir, evaluatorOptions(settings))
}

//goland:noinspection GoBoolExpressions
func generatorSettingsSource() (*pkl.ModuleSource, error) {
	if Version == "development" {
		_, filename, _, ok := runtime.Caller(1)
		if !ok {
			return nil, fmt.Errorf("%w: could not get path to main Go source file", ErrRuntimeCallerFailure)
		}
		dirPath := filepath.Dir(filename)
		return pkl.FileSource(dirPath, "../../codegen/src/GeneratorSettings.pkl"), nil
	}
	return pkl.UriSource(fmt.Sprintf("package://pkg.pkl-lang.org/pkl-go/pkl.golang@%s#/GeneratorSettings.pkl", Version)), nil
}

func newGeneratorSettingsEvaluator(dirname string) (pkl.Evaluator, error) {
	if dirname != "" {
		return pkl.NewProjectEvaluator(context.Background(), dirname, pkl.PreconfiguredOptions)
	}
	return pkl.NewEvaluator(context.Background(), pkl.PreconfiguredOptions)
}

func newModuleSource(filename string) (*pkl.ModuleSource, error) {

	// case 1: filename is provided directly
	if filename != "" {
		return pkl.FileSource(filename), nil
	}

	// case 2: filename is empty, but generator-settings.pkl may exist
	if exists, existsErr := fileExists("generator-settings.pkl"); existsErr != nil {
		return nil, existsErr
	} else if exists {
		return pkl.FileSource("generator-settings.pkl"), nil
	}

	// case 3: use the default generator settings source
	return generatorSettingsSource()
}

// loadGeneratorSettings loads the settings for controlling code generation.
//
// Uses a Pkl evaluator which is separate from what's used for actually running codegen.
func loadGeneratorSettings(generatorSettingsFilename, projDirname string) (*generatorsettings.GeneratorSettings, error) {
	// normalize the project directory
	absProjDirname, absErr := filepath.Abs(projDirname)
	if absErr != nil {
		return nil, absErr
	}

	// get the project evaluator
	evaluator, evalErr := newGeneratorSettingsEvaluator(absProjDirname)
	if evalErr != nil {
		return nil, evalErr
	}

	// get the module source
	source, sourceErr := newModuleSource(generatorSettingsFilename)
	if sourceErr != nil {
		return nil, sourceErr
	}

	// wrap this all together into an instance of GeneratorSettings
	return generatorsettings.Load(context.Background(), evaluator, source)
}

// handleFatalError handles an error that should end the program.
//
// When a non-zero error value is provided, we exit with a non-zero error code and an error message is written to
// stderr.
func handleFatalError(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "fatal error: %s\n", err)
		os.Exit(1)
	}
}

// unwrapValue "unwraps" the output of a function that returns a value and an error.
//
// It does this by assuming that any error which comes out of said function can be considered "fatal", in which case
// the program will immediately exit with an error message.
func unwrapValue[T any](t T, err error) T {
	handleFatalError(err)
	return t
}

func main() {
	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
}
