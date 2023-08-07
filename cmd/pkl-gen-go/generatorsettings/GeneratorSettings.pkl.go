// Code generated from Pkl module `temp.pkl.golang.GeneratorSettings`. DO NOT EDIT.
package generatorsettings

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

// Settings to configure Go code generation for Pkl files.
type GeneratorSettings struct {
	// Mapping of Pkl module names to their respective Go package name.
	//
	// Example:
	//
	// ```
	// packageMappings {
	//   ["name.of.my.Module"] = "github.com/myteam/myorg/foo"
	// }
	// ```
	PackageMappings map[string]string `pkl:"packageMappings"`

	// The base path for module output.
	//
	// This determines the relative path that generated Go source code will be written to.
	// For example, a base path of `"github.com/foo/bar"` means that the Go files for package
	// `github.com/foo/bar/baz` will be written into a `baz` directory, relative to the current
	// working directory where codegen is executed.
	//
	// Any Go packages that are not prefixed with [basePath] are skipped from code generation.
	//
	// This is typically a Go module's name, i.e. the `module` clause within a `go.mod` file.
	//
	// If empty, writes the full package path to the current directory.
	BasePath string `pkl:"basePath"`

	// Additional struct tags to place on all properties.
	//
	// In addition to these tags, every property implicitly receives a `pkl:` struct tag.
	//
	// The placeholder `%{name}` gets substituted with the name of the Pkl property.
	//
	// Struct tags can also be configured on a per-property basis using the [go.Field] annotation.
	//
	// Example:
	//
	// ```pkl
	// structTags {
	//   ["json"] = "%{name},omitempty"
	// }
	// ```
	StructTags map[string]string `pkl:"structTags"`

	// The path to the Pkl code generator script.
	//
	// This is an internal setting that is used for testing purposes when developing the code
	// generator.
	GeneratorScriptPath string `pkl:"generatorScriptPath"`

	// URI patterns that determine which modules can be loaded and evaluated.
	//
	// This corresponds to the `--allowed-modules` flag in the Pkl CLI.
	AllowedModules []string `pkl:"allowedModules"`

	// URI patterns that determine which external resources can be read.
	//
	// This corresponds to the `--allowed-resources` flag in the Pkl CLI.
	AllowedResources []string `pkl:"allowedResources"`

	// Print out the names of the files that will be generated, but skip writing anything to disk.
	DryRun bool `pkl:"dryRun"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a GeneratorSettings
func LoadFromPath(ctx context.Context, path string) (ret *GeneratorSettings, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a GeneratorSettings
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*GeneratorSettings, error) {
	var ret GeneratorSettings
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
