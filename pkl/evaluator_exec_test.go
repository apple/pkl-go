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
	"context"
	"net/url"
	"os"
	"path/filepath"
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

func TestEvaluatorManager_NewEvaluatorStartFailure(t *testing.T) {
	manager := NewEvaluatorManagerWithCommand([]string{filepath.Join(t.TempDir(), "missing-pkl")})

	ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions)
	assert.Nil(t, ev)
	assert.ErrorIs(t, err, os.ErrNotExist)

	// Closing the manager previously waited for the process that never started, and then
	// panicked when trying to kill it.
	assert.NoError(t, manager.Close())
}

func TestEvaluatorManager_NewEvaluatorAfterStartFailure(t *testing.T) {
	pklExec, hasPklExec := os.LookupEnv("PKL_EXEC")
	t.Setenv("PKL_EXEC", filepath.Join(t.TempDir(), "missing-pkl"))
	manager := NewEvaluatorManager()
	defer func() {
		assert.NoError(t, manager.Close())
	}()

	_, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions)
	assert.ErrorIs(t, err, os.ErrNotExist)

	if hasPklExec {
		t.Setenv("PKL_EXEC", pklExec)
	} else {
		assert.NoError(t, os.Unsetenv("PKL_EXEC"))
	}

	// A failed start must not close the manager; it can start Pkl once the executable is available.
	ev, err := manager.NewEvaluator(context.Background(), PreconfiguredOptions)
	if assert.NoError(t, err) {
		out, err := ev.EvaluateOutputText(context.Background(), TextSource("foo = 1"))
		assert.NoError(t, err)
		assert.Equal(t, "foo = 1\n", out)
	}
}
