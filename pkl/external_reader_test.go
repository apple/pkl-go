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

package pkl

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

const externalReaderTest1 = `
import "pkl:test"

fib5 = read("fib:5").text.toInt()
fib10 = read("fib:10").text.toInt()
fib100 = read("fib:20").text.toInt()

fibErrA = test.catch(() -> read("fib:%20"))
fibErrB = test.catch(() -> read("fib:abc"))
fibErrC = test.catch(() -> read("fib:-10"))
`

func TestExternalReaderE2E(t *testing.T) {
	manager := NewEvaluatorManager()
	version, err := manager.(*evaluatorManager).getVersion()
	if err != nil {
		t.Fatal(err)
	}
	if pklVersion0_27.isGreaterThan(version) {
		t.SkipNow()
	}

	tempDir := t.TempDir()
	writeFile(t, tempDir+"/test.pkl", externalReaderTest1)

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("can't find caller")
	}
	projectRoot := filepath.Join(filepath.Dir(filename), "../cmd/internal/test-external-reader/test-external-reader.go")

	evaluator, err := manager.NewEvaluator(
		context.Background(),
		PreconfiguredOptions,
		WithExternalResourceReader("fib", ExternalReader{
			Executable: "go",
			Arguments:  []string{"run", projectRoot},
		}),
	)
	if !assert.NoError(t, err) {
		return
	}

	output, err := evaluator.EvaluateOutputText(context.Background(), FileSource(tempDir+"/test.pkl"))
	assert.NoError(t, err)
	assert.Equal(t, output, `fib5 = 3
fib10 = 34
fib100 = 4181
fibErrA = "I/O error reading resource `+"`fib:%20`"+`. IOException: input uri must be in format fib:<positive integer>: non-positive value"
fibErrB = "I/O error reading resource `+"`fib:abc`"+`. IOException: input uri must be in format fib:<positive integer>: non-positive value"
fibErrC = "I/O error reading resource `+"`fib:-10`"+`. IOException: input uri must be in format fib:<positive integer>: non-positive value"
`)
}
