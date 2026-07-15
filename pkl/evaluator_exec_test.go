//===----------------------------------------------------------------------===//
// Copyright © 2024-2026 Apple Inc. and the Pkl project authors. All rights reserved.
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
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProjectEvaluatorRejectsFileURLWithoutPath(t *testing.T) {
	for _, raw := range []string{"file://.", "file:.", "file://localhost"} {
		u, err := url.Parse(raw)
		assert.NoError(t, err)

		// A file URI without a path previously reached Pkl and failed with an
		// opaque PklBugException; it should now fail fast with an actionable
		// error before any evaluator process is started.
		ev, err := NewProjectEvaluator(context.Background(), u)
		assert.Nil(t, ev, "expected no evaluator for %q", raw)
		assert.ErrorContains(t, err, "invalid file URI", "for %q", raw)
	}
}
