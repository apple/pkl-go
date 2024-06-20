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
	"context"
	"embed"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed test_fixtures/testfs/*
var testFs embed.FS

func setupProject(t *testing.T) string {
	tempDir := t.TempDir()
	_ = os.WriteFile(tempDir+"/PklProject", []byte(`
amends "pkl:Project"

dependencies {
  ["uri"] { uri = "package://pkg.pkl-lang.org/pkl-pantry/pkl.experimental.uri@1.0.0" }
}
`), 0o644)
	_ = os.WriteFile(tempDir+"/PklProject.deps.json", []byte(`
{
  "schemaVersion": 1,
  "resolvedDependencies": {
    "package://pkg.pkl-lang.org/pkl-pantry/pkl.experimental.uri@1": {
      "type": "remote",
      "uri": "projectpackage://pkg.pkl-lang.org/pkl-pantry/pkl.experimental.uri@1.0.0",
      "checksums": {
        "sha256": "12a42da6a2933a802cc79cea7f5541513b5106070ca5f1236009ebefeb3d81b3"
      }
    }
  }
}
`), 0o644)
	_ = os.WriteFile(tempDir+"/main.pkl", []byte(`
import "@uri/URI.pkl"

uri = URI.parse("https://www.example.com").toString()
`), 0o644)
	return tempDir
}

func getOpenPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer listener.Close()
	addrStr := listener.Addr().String()
	parts := strings.Split(addrStr, ":")
	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		panic(err)
	}
	return port
}

func TestEvaluator(t *testing.T) {
	manager := NewEvaluatorManager()

	projectDir := setupProject(t)

	t.Run("EvaluateOutputText", func(t *testing.T) {
		t.Parallel()
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions)
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource("foo { bar = 1 }"))
			assert.NoError(t, err)
			assert.Equal(t, "foo {\n  bar = 1\n}\n", out)
			out, err = ev.EvaluateOutputText(context.Background(), TextSource("bar { baz = 2 }"))
			assert.NoError(t, err)
			assert.Equal(t, "bar {\n  baz = 2\n}\n", out)
			assert.NoError(t, ev.Close())
		}
	})

	t.Run("EvaluateOutputText - output format", func(t *testing.T) {
		t.Parallel()
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, func(options *EvaluatorOptions) {
			options.OutputFormat = "yaml"
		})
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource("foo { bar = 1 }"))
			assert.NoError(t, err)
			assert.Equal(t, "foo:\n  bar: 1\n", out)
		}
	})

	t.Run("EvaluateOutputFiles", func(t *testing.T) {
		t.Parallel()
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions)
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputFiles(context.Background(), TextSource(`output {
  files {
    ["foo.txt"] { text = "foo text" }
    ["bar.txt"] { text = "bar text" }
  }
}`))
			assert.NoError(t, err)
			assert.Equal(t, map[string]string{
				"foo.txt": "foo text",
				"bar.txt": "bar text",
			}, out)
			assert.NoError(t, ev.Close())
		}
	})

	t.Run("EvaluateModule", func(t *testing.T) {
		t.Parallel()
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions)
		if assert.NoError(t, err) {
			type MyModule struct {
				Foo string `pkl:"foo"`
				Bar int    `pkl:"bar"`
			}
			var m MyModule
			err = ev.EvaluateModule(context.Background(), TextSource(`
foo: String = "foo"
bar: Int = 5
`), &m)
			assert.NoError(t, err)
			assert.Equal(t, MyModule{Foo: "foo", Bar: 5}, m)
		}
	})

	t.Run("custom logger", func(t *testing.T) {
		t.Parallel()
		s := &stubLogger{}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, func(options *EvaluatorOptions) {
			options.Logger = s
		})
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource("foo { bar = trace(\"bar\") }"))
			assert.NoError(t, err)
			assert.Equal(t, "foo {\n  bar = \"bar\"\n}\n", out)
			if assert.Len(t, s.traces, 1) {
				assert.Equal(t, s.traces[0], `"bar" = "bar"`)
			}
			assert.NoError(t, ev.Close())
		}
	})

	t.Run("custom resource reader", func(t *testing.T) {
		t.Parallel()
		reader := &virtualResourceReader{
			scheme: "flintstone",
			read: func(u url.URL) ([]byte, error) {
				return []byte("Fred Flintstone"), nil
			},
		}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, WithResourceReader(reader))
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource(`foo = read("flintstone:fred").text`))
			assert.NoError(t, err)
			assert.Equal(t, "foo = \"Fred Flintstone\"\n", out)
			assert.NoError(t, ev.Close())
		}
	})

	t.Run("custom resource reader error", func(t *testing.T) {
		t.Parallel()
		reader := &virtualResourceReader{
			scheme: "flintstone",
			read: func(url url.URL) ([]byte, error) {
				return nil, fmt.Errorf("cannot find resource %s", &url)
			},
		}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, WithResourceReader(reader))
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource(`foo = read("flintstone:fred").text`))
			assert.Empty(t, out)
			assert.Error(t, err)
			assert.IsType(t, &EvalError{}, err)
			assert.NoError(t, ev.Close())
		}
	})

	t.Run("custom resource reader: globbing", func(t *testing.T) {
		t.Parallel()
		reader := &virtualResourceReader{
			scheme: "flintstone",
			read: func(u url.URL) ([]byte, error) {
				switch u.Opaque {
				case "barney":
					return []byte("gumble"), nil
				case "wilma":
					return []byte("wilma"), nil
				default:
					return []byte("something else"), nil
				}
			},
			listElements: func(u url.URL) ([]PathElement, error) {
				return []PathElement{
					NewPathElement("barney", false),
					NewPathElement("wilma", false),
					NewPathElement("fred", false),
				}, nil
			},
			isGlobbable: true,
		}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, WithResourceReader(reader))
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource(`flintstones = read*("flintstone:*")`))
			assert.Nil(t, err)
			assert.Equal(t, `flintstones {
  ["flintstone:barney"] {
    uri = "flintstone:barney"
    text = "gumble"
    base64 = "Z3VtYmxl"
  }
  ["flintstone:fred"] {
    uri = "flintstone:fred"
    text = "something else"
    base64 = "c29tZXRoaW5nIGVsc2U="
  }
  ["flintstone:wilma"] {
    uri = "flintstone:wilma"
    text = "wilma"
    base64 = "d2lsbWE="
  }
}
`, out)
			assert.NoError(t, ev.Close())
		}
	})

	t.Run("custom resource reader: glob error", func(t *testing.T) {
		t.Parallel()
		reader := &virtualResourceReader{
			scheme: "flintstone",
			read: func(u url.URL) ([]byte, error) {
				return nil, nil
			},
			listElements: func(u url.URL) ([]PathElement, error) {
				return nil, fmt.Errorf("something went wrong")
			},
			isGlobbable: true,
		}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, WithResourceReader(reader))
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource(`flintstones = read*("flintstone:*")`))
			assert.Empty(t, out)
			assert.Error(t, err, "IOException: something went wrong")
			assert.Zero(t, out)
			assert.NoError(t, ev.Close())
		}
	})

	t.Run("custom module reader", func(t *testing.T) {
		t.Parallel()
		reader := &virtualModuleReader{
			scheme: "flintstone",
			read: func(u url.URL) (string, error) {
				return `foo = 1`, nil
			},
		}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, WithModuleReader(reader))
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource(`result = import("flintstone:fred").foo`))
			assert.NoError(t, err)
			assert.Equal(t, "result = 1\n", out)
			assert.NoError(t, ev.Close())
		}
	})

	t.Run("custom module reader error", func(t *testing.T) {
		t.Parallel()
		reader := &virtualModuleReader{
			scheme: "flintstone",
			read: func(u url.URL) (string, error) {
				return "", fmt.Errorf("no idea where %s is", &u)
			},
		}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, WithModuleReader(reader))
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource(`result = import("flintstone:fred").foo`))
			assert.Empty(t, out)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "no idea where flintstone:fred is")
			assert.NoError(t, ev.Close())
		}
	})

	t.Run("custom module reader: triple-dot imports", func(t *testing.T) {
		t.Parallel()
		reader := &virtualModuleReader{
			scheme:              "flintstone",
			isGlobbable:         true,
			hasHierarchicalUris: true,
			isLocal:             true,
			read: func(u url.URL) (string, error) {
				switch u.Path {
				case "/foo/bar/baz.pkl":
					return `res = import("...")`, nil
				case "/foo/baz.pkl":
					return "", errors.New("not here")
				case "/baz.pkl":
					return "bar = 1", nil
				default:
					t.FailNow()
				}
				return "", nil
			},
		}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, WithModuleReader(reader))
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), UriSource("flintstone:/foo/bar/baz.pkl"))
			assert.NoError(t, err)
			assert.Equal(t, `res {
  bar = 1
}
`, out)
		}
	})

	t.Run("custom module reader: globbing", func(t *testing.T) {
		t.Parallel()
		reader := &virtualModuleReader{
			scheme:              "flintstone",
			isGlobbable:         true,
			hasHierarchicalUris: true,
			read: func(u url.URL) (string, error) {
				switch u.Path {
				case "/foo.pkl":
					return `res = 1`, nil
				case "/bar.pkl":
					return "res = 2", nil
				default:
					t.FailNow()
				}
				return "", nil
			},
			listElements: func(u url.URL) ([]PathElement, error) {
				assert.Equal(t, "/", u.Path)
				return []PathElement{
					NewPathElement("foo.pkl", false),
					NewPathElement("bar.pkl", false),
				}, nil
			},
		}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, WithModuleReader(reader))
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource(`res = import*("flintstone:/**.pkl")`))
			assert.NoError(t, err)
			assert.Equal(t, `res {
  ["flintstone:/bar.pkl"] {
    res = 2
  }
  ["flintstone:/foo.pkl"] {
    res = 1
  }
}
`, out)
		}
	})

	t.Run("custom module reader: glob error", func(t *testing.T) {
		t.Parallel()
		reader := &virtualModuleReader{
			scheme:              "flintstone",
			isGlobbable:         true,
			hasHierarchicalUris: true,
			read: func(u url.URL) (string, error) {
				t.FailNow()
				return "", nil
			},
			listElements: func(u url.URL) ([]PathElement, error) {
				return nil, fmt.Errorf("i failed")
			},
		}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, WithModuleReader(reader))
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), TextSource(`res = import*("flintstone:/**.pkl")`))
			assert.Error(t, err)
			assert.Zero(t, out)
		}
	})

	t.Run("custom fs", func(t *testing.T) {
		t.Parallel()
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, WithFs(testFs, "testfs"))
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), UriSource("testfs:/test_fixtures/testfs/person.pkl"))
			assert.NoError(t, err)
			assert.Equal(t, `name = "Barney"
age = 43
`, out)
			out, err = ev.EvaluateOutputText(context.Background(), UriSource("testfs:/test_fixtures/testfs/subdir/person.pkl"))
			assert.NoError(t, err)
			assert.Equal(t, `name = "Fred"
age = 43
`, out)
			out, err = ev.EvaluateOutputText(context.Background(), TextSource(`result = import*("testfs:/**.pkl")`))
			assert.NoError(t, err)
			assert.Equal(t, `result {
  ["testfs:/test_fixtures/testfs/person.pkl"] {
    name = "Barney"
    age = 43
  }
  ["testfs:/test_fixtures/testfs/subdir/person.pkl"] {
    name = "Fred"
    age = 43
  }
}
`, out)
		}
	})

	t.Run("EvaluatorManager.NewProjectEvaluator", func(t *testing.T) {
		// TODO(oss): re-enable this test after repos are public
		t.SkipNow()
		ev, err := manager.NewProjectEvaluator(context.Background(), projectDir, PreconfiguredOptions)
		if assert.NoError(t, err) {
			out, err := ev.EvaluateOutputText(context.Background(), FileSource(projectDir, "main.pkl"))
			assert.NoError(t, err)
			assert.Equal(t, "uri = \"https://www.example.com\"\n", out)
		}
	})

	t.Run("evaluate after close", func(t *testing.T) {
		t.Parallel()
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions)
		if assert.NoError(t, err) {
			assert.NoError(t, ev.Close())
			out, err := ev.EvaluateOutputText(context.Background(), TextSource("foo = 1"))
			assert.Empty(t, out)
			assert.Error(t, err, "evaluator is closed")
		}
	})

	t.Run("concurrent evaluations", func(t *testing.T) {
		t.Parallel()
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions)
		if err != nil {
			t.Fatal(err)
		}
		ch := make(chan string)
		for i := 0; i < 5; i++ {
			go func(j int) {
				res, _ := ev.EvaluateOutputText(context.Background(), TextSource(fmt.Sprintf("foo = %d", j)))
				ch <- res
			}(i)
		}

		var outputs []string
		for i := 0; i < 5; i++ {
			outputs = append(outputs, <-ch)
		}
		expected := []string{
			"foo = 0\n",
			"foo = 1\n",
			"foo = 2\n",
			"foo = 3\n",
			"foo = 4\n",
		}
		sort.Strings(outputs)
		assert.Equal(t, expected, outputs)
	})

	t.Run("concurrent new evaluators", func(t *testing.T) {
		t.Parallel()
		cherr := make(chan error)
		ch := make(chan Evaluator)
		for i := 0; i < 5; i++ {
			go func() {
				ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions)
				if err != nil {
					cherr <- err
				}
				ch <- ev
			}()
		}
		for i := 0; i < 5; i++ {
			select {
			case err := <-cherr:
				t.Fatal(err)
			case ev := <-ch:
				err := ev.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	})

	t.Run("custom proxy options", func(t *testing.T) {
		version, err := manager.(*evaluatorManager).getVersion()
		if err != nil {
			t.Fatal(err)
		}
		if pklVersion0_26.isGreaterThan(version) {
			t.SkipNow()
		}
		ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions, func(options *EvaluatorOptions) {
			options.Http = &Http{
				Proxy: &Proxy{
					Address: fmt.Sprintf("http://localhost:%d", getOpenPort()),
				},
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ev.EvaluateOutputText(context.Background(), UriSource("https://example.com"))
		assert.ErrorContains(t, err, "ConnectException: Error connecting to host `example.com`")
	})

	t.Run("custom proxy options errors on Pkl 0.25", func(t *testing.T) {
		version, err := manager.(*evaluatorManager).getVersion()
		if err != nil {
			t.Fatal(err)
		}
		if version.isGreaterThan(pklVersion0_25) {
			t.SkipNow()
		}
		_, err = manager.NewEvaluator(context.Background(), PreconfiguredOptions, func(options *EvaluatorOptions) {
			options.Http = &Http{
				Proxy: &Proxy{
					Address: fmt.Sprintf("http://localhost:%d", getOpenPort()),
				},
			}
		})
		assert.ErrorContains(t, err, "http options are not supported on Pkl versions lower than 0.26")
	})

	t.Cleanup(func() {
		assert.NoError(t, manager.Close())
	})
}

func TestNewProjectEvaluator(t *testing.T) {
	// TODO(oss): re-enable this test after repos are public
	t.SkipNow()
	projectDir := setupProject(t)
	ev, err := NewProjectEvaluator(context.Background(), projectDir, PreconfiguredOptions)
	if assert.NoError(t, err) {
		out, err := ev.EvaluateOutputText(context.Background(), FileSource(projectDir, "main.pkl"))
		assert.NoError(t, err)
		assert.Equal(t, "uri = \"https://www.example.com\"\n", out)
	}
}

type stubLogger struct {
	traces []string
	warns  []string
}

func (s *stubLogger) Trace(message string, _ string) {
	s.traces = append(s.traces, message)
}

func (s *stubLogger) Warn(message string, _ string) {
	s.warns = append(s.warns, message)
}

var _ Logger = (*stubLogger)(nil)

type virtualResourceReader struct {
	scheme              string
	isGlobbable         bool
	hasHierarchicalUris bool
	read                func(u url.URL) ([]byte, error)
	listElements        func(u url.URL) ([]PathElement, error)
}

func (v virtualResourceReader) IsGlobbable() bool {
	return v.isGlobbable
}

func (v virtualResourceReader) HasHierarchicalUris() bool {
	return v.hasHierarchicalUris
}

func (v virtualResourceReader) ListElements(u url.URL) ([]PathElement, error) {
	return v.listElements(u)
}

func (v virtualResourceReader) Scheme() string {
	return v.scheme
}

func (v virtualResourceReader) Read(u url.URL) ([]byte, error) {
	return v.read(u)
}

var _ ResourceReader = (*virtualResourceReader)(nil)

type virtualModuleReader struct {
	scheme              string
	isGlobbable         bool
	isLocal             bool
	hasHierarchicalUris bool
	read                func(u url.URL) (string, error)
	listElements        func(u url.URL) ([]PathElement, error)
}

func (v virtualModuleReader) IsGlobbable() bool {
	return v.isGlobbable
}

func (v virtualModuleReader) HasHierarchicalUris() bool {
	return v.hasHierarchicalUris
}

func (v virtualModuleReader) IsLocal() bool {
	return v.isLocal
}

func (v virtualModuleReader) ListElements(u url.URL) ([]PathElement, error) {
	return v.listElements(u)
}

func (v virtualModuleReader) Scheme() string {
	return v.scheme
}

func (v virtualModuleReader) Read(u url.URL) (string, error) {
	return v.read(u)
}

var _ ModuleReader = (*virtualModuleReader)(nil)

func TestEvaluateAfterClose(t *testing.T) {
	manager := NewEvaluatorManager()
	err := manager.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = manager.NewEvaluator(context.Background(), PreconfiguredOptions)
	assert.Error(t, err, "evaluator is closed")
}
