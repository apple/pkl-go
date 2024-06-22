package pkl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var semverPattern = regexp.MustCompile(`(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`)

var numericIdentifer = regexp.MustCompile(`^(0|[1-9]\d*)$`)

//goland:noinspection GoSnakeCaseUsage
var pklVersion0_25 = mustParseSemver("0.25.0")

//goland:noinspection GoSnakeCaseUsage
var pklVersion0_26 = mustParseSemver("0.26.0")

type semver struct {
	major                 int
	minor                 int
	patch                 int
	prerelease            string
	build                 string
	prereleaseIdentifiers []prereleaseIdentifier
}

type prereleaseIdentifier struct {
	numericId      int
	alphaNumericId string
}

func (i prereleaseIdentifier) compareTo(other prereleaseIdentifier) int {
	if i.alphaNumericId != "" {
		return strings.Compare(i.alphaNumericId, other.alphaNumericId)
	} else {
		return compareInt(i.numericId, other.numericId)
	}
}

func (s *semver) getPrereleaseIdentifiers() []prereleaseIdentifier {
	if s.prerelease == "" {
		return nil
	}
	if len(s.prereleaseIdentifiers) > 0 {
		return s.prereleaseIdentifiers
	}
	identifiers := strings.Split(s.prerelease, ".")
	prereleaseIdentifiers := make([]prereleaseIdentifier, len(identifiers))
	for i, str := range identifiers {
		if numericIdentifer.MatchString(str) {
			// guaranteed to succeed
			num, _ := strconv.Atoi(str)
			prereleaseIdentifiers[i] = prereleaseIdentifier{numericId: num}
		} else {
			prereleaseIdentifiers[i] = prereleaseIdentifier{alphaNumericId: str}
		}
	}
	s.prereleaseIdentifiers = prereleaseIdentifiers
	return s.prereleaseIdentifiers
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func (s *semver) compareToString(other string) (int, error) {
	otherVersion, err := parseSemver(other)
	if err != nil {
		return 0, err
	}
	return s.compareTo(otherVersion), nil
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
	comparison = compareInt(s.patch, other.patch)
	if comparison != 0 {
		return comparison
	}
	ids1 := s.getPrereleaseIdentifiers()
	ids2 := other.getPrereleaseIdentifiers()
	for i := 0; i < min(len(ids1), len(ids2)); i++ {
		cmp := ids1[i].compareTo(ids2[i])
		if cmp != 0 {
			return cmp
		}
	}
	return compareInt(len(ids1), len(ids2))
}

func (s *semver) isGreaterThan(other *semver) bool {
	return s.compareTo(other) > 0
}

func (s *semver) String() string {
	res := fmt.Sprintf("%d.%d.%d", s.major, s.minor, s.patch)
	if s.prerelease == "" && s.build == "" {
		return res
	}
	if s.prerelease != "" {
		res += "-" + s.prerelease
	}
	if s.build != "" {
		res += "+" + s.build
	}
	return res
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
	if len(matched) < 6 {
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
