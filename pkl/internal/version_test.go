//===----------------------------------------------------------------------===//
// Copyright © 2025 Apple Inc. and the Pkl project authors. All rights reserved.
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

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func compareVersions(v1 string, v2 string) int {
	semverV1 := MustParseSemver(v1)
	semverV2 := MustParseSemver(v2)
	return semverV1.CompareTo(semverV2)
}

func TestCompareSemverVersions(t *testing.T) {
	assert.Equal(t, -1, compareVersions("1.0.0", "2.0.0"))
	assert.Equal(t, -1, compareVersions("1.0.0", "1.0.1"))
	assert.Equal(t, -1, compareVersions("1.0.0", "1.1.0"))
	assert.Equal(t, -1, compareVersions("1.0.5", "1.5.0"))
	assert.Equal(t, 0, compareVersions("1.0.0", "1.0.0"))
	assert.Equal(t, 0, compareVersions("1.1.0", "1.1.0"))
	assert.Equal(t, 0, compareVersions("5.1.0", "5.1.0"))
	assert.Equal(t, 1, compareVersions("2.0.0", "1.0.0"))
	assert.Equal(t, 1, compareVersions("2.0.0", "0.2.0"))
	assert.Equal(t, 1, compareVersions("2.0.0", "0.0.2"))
	assert.Equal(t, 1, compareVersions("2.0.0", "0.0.15"))
	assert.Equal(t, -1, compareVersions("2.0.0-alpha", "2.0.0-beta"))
	assert.Equal(t, 1, compareVersions("2.0.0-alpha", "2.0.0-aaa"))
	assert.Equal(t, 0, compareVersions("2.0.0-alpha", "2.0.0-alpha"))
	assert.Equal(t, 0, compareVersions("2.0.0-1.2.3", "2.0.0-1.2.3"))
	assert.Equal(t, 1, compareVersions("2.0.0-1.2.3", "2.0.0-1.2.2"))
	assert.Equal(t, 0, compareVersions("2.0.0-a.b.3", "2.0.0-a.b.3"))
	assert.Equal(t, 1, compareVersions("2.0.0-1.2.3.4", "2.0.0-1.2.3"))
	assert.Equal(t, 0, compareVersions("2.0.0+foo", "2.0.0+bar"))
}
