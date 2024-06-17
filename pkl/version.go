package pkl

import (
	"fmt"
	"strconv"
)

//goland:noinspection GoSnakeCaseUsage
var pklVersion0_25 = mustParseSemver("0.25.0")

//goland:noinspection GoSnakeCaseUsage
var pklVersion0_26 = mustParseSemver("0.26.0")

type semver struct {
	major      int
	minor      int
	patch      int
	prerelease string
	build      string
}

func compareInt(a, b int) int {
	if a == b {
		return 0
	}
	if a < b {
		return -1
	}
	return 1
}

func (s *semver) compareToString(other string) int {
	otherVersion, err := parseSemver(other)
	if err != nil {
		return 0
	}
	return s.compareTo(otherVersion)
}

// compareTo returns -1 if s < other, 1 if s > other, and 0 otherwise.
func (s *semver) compareTo(other *semver) int {
	comparison := compareInt(s.major, other.major)
	if comparison != 0 {
		return comparison
	}
	comparison = compareInt(s.minor, other.minor)
	if comparison != 0 {
		return comparison
	}
	// technically we should proceed to comparing prerelease versions, but we can skip
	// this part because we don't have a use-case for it.
	return compareInt(s.patch, other.patch)
}

func (s *semver) isGreaterThanString(other string) bool {
	return s.compareToString(other) == 1
}

func (s *semver) isLessThanOrEqualToString(other string) bool {
	parsed, err := parseSemver(other)
	if err != nil {
		return false
	}
	return s.compareTo(parsed) <= 0
}

func (s *semver) isGreaterThanOrEqualToString(other string) bool {
	return s.compareToString(other) >= 0
}

func (s *semver) String() string {
	return fmt.Sprintf("%d.%d.%d", s.major, s.minor, s.patch)
}

func mustParseSemver(s string) *semver {
	parsed, err := parseSemver(s)
	if err != nil {
		panic(err)
	}
	return parsed
}

func parseSemver(s string) (*semver, error) {
	matched := semverPattern.FindStringSubmatch(s)
	if len(matched) < 5 {
		return nil, fmt.Errorf("failed to parse %s as semver", s)
	}
	major, err := strconv.Atoi(matched[1])
	if err != nil {
		return nil, err
	}
	minor, err := strconv.Atoi(matched[2])
	if err != nil {
		return nil, err
	}
	patch, err := strconv.Atoi(matched[3])
	if err != nil {
		return nil, err
	}
	return &semver{
		major:      major,
		minor:      minor,
		patch:      patch,
		prerelease: matched[4],
		build:      matched[5],
	}, nil
}
