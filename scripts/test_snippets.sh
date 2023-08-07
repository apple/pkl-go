#!/usr/bin/env bash
# ===----------------------------------------------------------------------===//
# Copyright © 2024 Apple Inc. and the Pkl project authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ===----------------------------------------------------------------------===//

set -e

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
SNIPPET_TEST_DIR="$SCRIPT_DIR/codegen/snippet-tests"

rm -rf "$SNIPPET_TEST_DIR/output"

go run cmd/internal/gen-snippets/gen-snippets.go

diff=$(git diff codegen/snippet-tests/)

if [[ -n "$diff" ]]; then
  echo "Error: Snippet tests contains changes!"
  echo "$diff"
  exit 1
fi
