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
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var SemverPattern = regexp.MustCompile(`(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`)

var numericIdentifier = regexp.MustCompile(`^(0|[1-9]\d*)$`)

//goland:noinspection GoSnakeCaseUsage
var PklVersion0_25 = MustParseSemver("0.25.0")

//goland:noinspection GoSnakeCaseUsage
var PklVersion0_26 = MustParseSemver("0.26.0")

//goland:noinspection GoSnakeCaseUsage
var PklVersion0_27 = MustParseSemver("0.27.0")

//goland:noinspection GoSnakeCaseUsage
var PklVersion0_28 = MustParseSemver("0.28.0")

//goland:noinspection GoSnakeCaseUsage
var PklVersion0_29 = MustParseSemver("0.29.0")

//goland:noinspection GoSnakeCaseUsage
var PklVersion0_30 = MustParseSemver("0.30.0")

type Semver struct {
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
	}
	return compareInt(i.numericId, other.numericId)
}

func (s *Semver) getPrereleaseIdentifiers() []prereleaseIdentifier {
	if s.prerelease == "" {
		return nil
	}
	if len(s.prereleaseIdentifiers) > 0 {
		return s.prereleaseIdentifiers
	}
	identifiers := strings.Split(s.prerelease, ".")
	prereleaseIdentifiers := make([]prereleaseIdentifier, len(identifiers))
	for i, str := range identifiers {
		if numericIdentifier.MatchString(str) {
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

func (s *Semver) compareToString(other string) (int, error) {
	otherVersion, err := ParseSemver(other)
	if err != nil {
		return 0, err
	}
	return s.CompareTo(otherVersion), nil
}

// CompareTo returns -1 if s < other, 1 if s > other, and 0 otherwise.
func (s *Semver) CompareTo(other *Semver) int {
	if comparison := compareInt(s.major, other.major); comparison != 0 {
		return comparison
	}
	if comparison := compareInt(s.minor, other.minor); comparison != 0 {
		return comparison
	}
	// technically we should proceed to comparing prerelease versions, but we can skip
	// this part because we don't have a use-case for it.
	if comparison := compareInt(s.patch, other.patch); comparison != 0 {
		return comparison
	}
	ids1, ids2 := s.getPrereleaseIdentifiers(), other.getPrereleaseIdentifiers()
	// if one version is stable (no prerelease ids) then it is higher
	if len(ids1) == 0 || len(ids2) == 0 {
		return compareInt(len(ids2), len(ids1))
	}
	// otherwise, pair-wise compare each prelease id
	for i := 0; i < min(len(ids1), len(ids2)); i++ {
		if cmp := ids1[i].compareTo(ids2[i]); cmp != 0 {
			return cmp
		}
	}
	// fall back to more prerelease ids is higher
	return compareInt(len(ids1), len(ids2))
}

func (s *Semver) IsGreaterThan(other *Semver) bool {
	return s.CompareTo(other) > 0
}

func (s *Semver) IsLessThan(other *Semver) bool {
	return s.CompareTo(other) < 0
}

func (s *Semver) String() string {
	var builder strings.Builder
	_, err := fmt.Fprintf(&builder, "%d.%d.%d", s.major, s.minor, s.patch)
	if err != nil {
		// should never happen
		panic(err.Error())
	}
	if s.prerelease != "" {
		builder.WriteByte('-')
		builder.WriteString(s.prerelease)
	}
	if s.build != "" {
		builder.WriteByte('+')
		builder.WriteString(s.build)
	}
	return builder.String()
}

func MustParseSemver(s string) *Semver {
	parsed, err := ParseSemver(s)
	if err != nil {
		panic(err)
	}
	return parsed
}

func ParseSemver(s string) (*Semver, error) {
	matched := SemverPattern.FindStringSubmatch(s)
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
	return &Semver{
		major:      major,
		minor:      minor,
		patch:      patch,
		prerelease: matched[4],
		build:      matched[5],
	}, nil
}
