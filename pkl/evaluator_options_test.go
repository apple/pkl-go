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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

// decodeCreateEvaluator encodes the options as a CreateEvaluator message and decodes the payload
// into a generic map, so that tests can assert on the keys that are sent over the wire.
func decodeCreateEvaluator(t *testing.T, opts *EvaluatorOptions) map[string]any {
	t.Helper()
	packed, err := opts.toMessage().ToMsgPack()
	require.NoError(t, err)

	dec := msgpack.NewDecoder(bytes.NewReader(packed))
	n, err := dec.DecodeArrayLen()
	require.NoError(t, err)
	require.Equal(t, 2, n)
	_, err = dec.DecodeInt()
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, dec.Decode(&payload))
	return payload
}

func TestExternalReader_toMessage(t *testing.T) {
	reader := &ExternalReader{
		Executable: "reader",
		Arguments:  []string{"--flag"},
		WorkingDir: "/some/dir",
	}
	msg := reader.toMessage()
	assert.Equal(t, "reader", msg.Executable)
	assert.Equal(t, []string{"--flag"}, msg.Arguments)
	assert.Equal(t, "/some/dir", msg.WorkingDir)
}

func TestCreateEvaluator_externalReaderWorkingDir(t *testing.T) {
	payload := decodeCreateEvaluator(t, &EvaluatorOptions{
		ExternalModuleReaders: map[string]ExternalReader{
			"mod": {Executable: "reader", WorkingDir: "/module/dir"},
		},
		ExternalResourceReaders: map[string]ExternalReader{
			"res": {Executable: "reader", WorkingDir: "/resource/dir"},
		},
	})

	moduleReader := payload["externalModuleReaders"].(map[string]any)["mod"].(map[string]any)
	assert.Equal(t, "reader", moduleReader["executable"])
	assert.Equal(t, "/module/dir", moduleReader["workingDir"])
	assert.NotContains(t, moduleReader, "WorkingDir")

	resourceReader := payload["externalResourceReaders"].(map[string]any)["res"].(map[string]any)
	assert.Equal(t, "/resource/dir", resourceReader["workingDir"])
}

func TestCreateEvaluator_externalReaderWithoutWorkingDirOmitsKey(t *testing.T) {
	payload := decodeCreateEvaluator(t, &EvaluatorOptions{
		ExternalModuleReaders: map[string]ExternalReader{
			"mod": {Executable: "reader"},
		},
	})

	moduleReader := payload["externalModuleReaders"].(map[string]any)["mod"].(map[string]any)
	assert.NotContains(t, moduleReader, "workingDir")
	assert.NotContains(t, moduleReader, "WorkingDir")
}

func TestCreateEvaluator_dependencyChecksumsUseSha256Key(t *testing.T) {
	payload := decodeCreateEvaluator(t, &EvaluatorOptions{
		ProjectBaseURI: "file:///project",
		DeclaredProjectDependencies: &ProjectDependencies{
			RemoteDependencies: map[string]*ProjectRemoteDependency{
				"dep": {
					PackageUri: "package://example.com/dep@1.0.0",
					Checksums:  &Checksums{Sha256: "abc123"},
				},
			},
		},
	})

	dependencies := payload["project"].(map[string]any)["dependencies"].(map[string]any)
	checksums := dependencies["dep"].(map[string]any)["checksums"].(map[string]any)
	assert.Equal(t, map[string]any{"sha256": "abc123"}, checksums)
}
