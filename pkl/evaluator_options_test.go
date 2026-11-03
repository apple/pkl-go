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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
}
