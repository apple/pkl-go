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

package pkl

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

func TestWithHttpHeaders(t *testing.T) {
	t.Parallel()
	for name, setup := range map[string]func(*EvaluatorOptions){
		"fresh options":         func(*EvaluatorOptions) {},
		"preconfigured options": PreconfiguredOptions,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			opts := EvaluatorOptions{}
			setup(&opts)
			require.NotPanics(t, func() {
				WithHttpHeaders("**", http.Header{"X-Test": {"one"}})(&opts)
			})
			require.NotNil(t, opts.Http)
			assert.Equal(t, http.Header{"X-Test": {"one"}}, opts.Http.Headers["**"])
		})
	}
	t.Run("preserves existing HTTP settings", func(t *testing.T) {
		t.Parallel()
		proxy := &Proxy{Address: "http://localhost:8080"}
		existing := &Http{
			CaCertificates: []byte("certificate"),
			Proxy:          proxy,
			Rewrites:       map[string]string{"https://from/": "https://to/"},
		}
		opts := EvaluatorOptions{Http: existing}
		WithHttpHeaders("**", http.Header{"X-Test": {"one"}})(&opts)
		assert.Same(t, existing, opts.Http)
		assert.Same(t, proxy, opts.Http.Proxy)
		assert.Equal(t, []byte("certificate"), opts.Http.CaCertificates)
		assert.Equal(t, map[string]string{"https://from/": "https://to/"}, opts.Http.Rewrites)
		assert.Equal(t, http.Header{"X-Test": {"one"}}, opts.Http.Headers["**"])
	})
	t.Run("appends repeated patterns and retains other patterns", func(t *testing.T) {
		t.Parallel()
		opts := EvaluatorOptions{}
		require.NotPanics(t, func() {
			WithHttpHeaders("**", http.Header{"X-Test": {"one"}})(&opts)
			WithHttpHeaders("https://example.com/**", http.Header{"X-Other": {"value"}})(&opts)
			WithHttpHeaders("**", http.Header{"X-Test": {"two"}, "X-New": {"value"}})(&opts)
		})
		assert.Equal(t, map[string]http.Header{
			"**":                     {"X-Test": {"one", "two"}, "X-New": {"value"}},
			"https://example.com/**": {"X-Other": {"value"}},
		}, opts.Http.Headers)
	})
	t.Run("can be reused for more evaluators", func(t *testing.T) {
		t.Parallel()
		headers := http.Header{"X-Test": {"one"}}
		withHeaders := []func(*EvaluatorOptions){
			WithHttpHeaders("**", headers),
			WithHttpHeaders("**", http.Header{"X-New": {"value"}}),
		}
		for range 2 {
			var opts EvaluatorOptions
			for _, f := range withHeaders {
				f(&opts)
			}
			assert.Equal(t, http.Header{"X-Test": {"one"}, "X-New": {"value"}}, opts.Http.Headers["**"])
		}
		assert.Equal(t, http.Header{"X-Test": {"one"}}, headers)
	})
}

func TestWithProjectEvaluatorSettings(t *testing.T) {
	t.Parallel()
	t.Run("applies modulePath and timeout", func(t *testing.T) {
		t.Parallel()
		project := &Project{
			ResolvedEvaluatorSettings: ProjectEvaluatorSettings{
				ModulePath: []string{"/project/lib", "/project/vendor.zip"},
				Timeout:    Duration{Value: 5, Unit: Minute},
			},
		}
		opts := EvaluatorOptions{}
		WithProjectEvaluatorSettings(project)(&opts)
		assert.Equal(t, []string{"/project/lib", "/project/vendor.zip"}, opts.ModulePaths)
		assert.Equal(t, int64(300), opts.TimeoutSeconds)
	})
	t.Run("rounds a sub-second timeout up", func(t *testing.T) {
		t.Parallel()
		project := &Project{
			ResolvedEvaluatorSettings: ProjectEvaluatorSettings{
				Timeout: Duration{Value: 1500, Unit: Millisecond},
			},
		}
		opts := EvaluatorOptions{}
		WithProjectEvaluatorSettings(project)(&opts)
		assert.Equal(t, int64(2), opts.TimeoutSeconds)
	})
	t.Run("keeps existing options when the project doesn't set them", func(t *testing.T) {
		t.Parallel()
		opts := EvaluatorOptions{ModulePaths: []string{"/existing"}, TimeoutSeconds: 10}
		WithProjectEvaluatorSettings(&Project{})(&opts)
		assert.Equal(t, []string{"/existing"}, opts.ModulePaths)
		assert.Equal(t, int64(10), opts.TimeoutSeconds)
	})
}

// decodeCreateEvaluator encodes the options as a CreateEvaluator message and decodes the payload
// into a generic map, so that tests can assert on the keys that are sent over the wire.
func decodeCreateEvaluator(t *testing.T, opts *EvaluatorOptions) map[string]any {
	t.Helper()
	packed, err := opts.toMessage().ToMsgPack()
	require.NoError(t, err)

	dec := msgpack.NewDecoder(bytes.NewReader(packed))
	n, err := dec.DecodeArrayLen()
	require.NoError(t, err)
	require.Equal(t, 2, n)
	_, err = dec.DecodeInt()
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, dec.Decode(&payload))
	return payload
}

func TestExternalReader_toMessage(t *testing.T) {
	reader := &ExternalReader{
		Executable: "reader",
		Arguments:  []string{"--flag"},
		WorkingDir: "/some/dir",
	}
	msg := reader.toMessage()
	assert.Equal(t, "reader", msg.Executable)
	assert.Equal(t, []string{"--flag"}, msg.Arguments)
	assert.Equal(t, "/some/dir", msg.WorkingDir)
}

func TestCreateEvaluator_externalReaderWorkingDir(t *testing.T) {
	payload := decodeCreateEvaluator(t, &EvaluatorOptions{
		ExternalModuleReaders: map[string]ExternalReader{
			"mod": {Executable: "reader", WorkingDir: "/module/dir"},
		},
		ExternalResourceReaders: map[string]ExternalReader{
			"res": {Executable: "reader", WorkingDir: "/resource/dir"},
		},
	})

	moduleReader := payload["externalModuleReaders"].(map[string]any)["mod"].(map[string]any)
	assert.Equal(t, "reader", moduleReader["executable"])
	assert.Equal(t, "/module/dir", moduleReader["workingDir"])
	assert.NotContains(t, moduleReader, "WorkingDir")

	resourceReader := payload["externalResourceReaders"].(map[string]any)["res"].(map[string]any)
	assert.Equal(t, "/resource/dir", resourceReader["workingDir"])
}

func TestCreateEvaluator_externalReaderWithoutWorkingDirOmitsKey(t *testing.T) {
	payload := decodeCreateEvaluator(t, &EvaluatorOptions{
		ExternalModuleReaders: map[string]ExternalReader{
			"mod": {Executable: "reader"},
		},
	})

	moduleReader := payload["externalModuleReaders"].(map[string]any)["mod"].(map[string]any)
	assert.NotContains(t, moduleReader, "workingDir")
	assert.NotContains(t, moduleReader, "WorkingDir")
}

func TestCreateEvaluator_dependencyChecksumsUseSha256Key(t *testing.T) {
	payload := decodeCreateEvaluator(t, &EvaluatorOptions{
		ProjectBaseURI: "file:///project",
		DeclaredProjectDependencies: &ProjectDependencies{
			RemoteDependencies: map[string]*ProjectRemoteDependency{
				"dep": {
					PackageUri: "package://example.com/dep@1.0.0",
					Checksums:  &Checksums{Sha256: "abc123"},
				},
			},
		},
	})

	dependencies := payload["project"].(map[string]any)["dependencies"].(map[string]any)
	checksums := dependencies["dep"].(map[string]any)["checksums"].(map[string]any)
	assert.Equal(t, map[string]any{"sha256": "abc123"}, checksums)
}
