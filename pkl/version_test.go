package pkl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func compareVersions(v1 string, v2 string) int {
	semverV1 := mustParseSemver(v1)
	semverV2 := mustParseSemver(v2)
	return semverV1.compareTo(semverV2)
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
