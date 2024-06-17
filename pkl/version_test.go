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
	assert.Equal(t, compareVersions("1.0.0", "2.0.0"), -1)
	assert.Equal(t, compareVersions("1.0.0", "1.0.1"), -1)
	assert.Equal(t, compareVersions("1.0.0", "1.1.0"), -1)
	assert.Equal(t, compareVersions("1.0.5", "1.5.0"), -1)
	assert.Equal(t, compareVersions("1.0.0", "1.0.0"), 0)
	assert.Equal(t, compareVersions("1.1.0", "1.1.0"), 0)
	assert.Equal(t, compareVersions("5.1.0", "5.1.0"), 0)
	assert.Equal(t, compareVersions("2.0.0", "1.0.0"), 1)
	assert.Equal(t, compareVersions("2.0.0", "0.2.0"), 1)
	assert.Equal(t, compareVersions("2.0.0", "0.0.2"), 1)
	assert.Equal(t, compareVersions("2.0.0", "0.0.15"), 1)
}
