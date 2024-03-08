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
package genpkl

import (
	"github.com/apple/pkl-go/pkl/test_fixtures/custom"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerator_GeneratePkl(t *testing.T) {
	tests := []struct {
		name      string
		generator Generator
		wantErr   bool
	}{
		{
			name: "customClass",
			generator: Generator{
				Modules: []interface{}{custom.CustomClasses{}},
				OutDir:  filepath.Join("..", "test_fixtures", "custom", "generated"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.generator.OutDir, "no Generator.OutDir")
			err := os.MkdirAll(tt.generator.OutDir, DefaultDirWritePermissions)
			require.NoError(t, err, "failed to make output directory %s", tt.generator.OutDir)

			err = tt.generator.GeneratePkl()
			if tt.wantErr {
				require.Error(t, err, "should have failed")
				return
			}
			require.NoError(t, err, "should not have failed")
		})
	}
}
