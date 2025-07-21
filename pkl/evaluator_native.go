//===----------------------------------------------------------------------===//
// Copyright © 2024-2025 Apple Inc. and the Pkl project authors. All rights reserved.
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

//go:build libpkl

package pkl

import (
	"context"
	"path/filepath"
)

// NewEvaluator returns an evaluator backed by a single EvaluatorManager.
// Its manager gets closed when the evaluator is closed.
//
// If creating multiple evaluators, prefer using EvaluatorManager.NewEvaluator instead,
// because it lessens the overhead of each successive evaluator.
func NewEvaluator(ctx context.Context, opts ...func(options *EvaluatorOptions)) (Evaluator, error) {
	manager := NewEvaluatorManager()
	ev, err := manager.NewEvaluator(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &simpleEvaluator{Evaluator: ev, manager: manager}, nil
}

// NewProjectEvaluator is an easy way to create an evaluator that is configured by the specified
// projectDir.
//
// It is similar to running the `pkl eval` or `pkl test` CLI command with a set `--project-dir`.
//
// When using project dependencies, they must first be resolved using the `pkl project resolve`
// CLI command.
func NewProjectEvaluator(ctx context.Context, projectDir string, opts ...func(options *EvaluatorOptions)) (Evaluator, error) {
	manager := NewEvaluatorManager()
	projectEvaluator, err := manager.NewEvaluator(ctx, opts...)
	if err != nil {
		return nil, err
	}
	defer projectEvaluator.Close()

	projectPath := filepath.Join(projectDir, "PklProject")
	project, err := LoadProjectFromEvaluator(ctx, projectEvaluator, projectPath)
	if err != nil {
		return nil, err
	}
	newOpts := []func(options *EvaluatorOptions){
		WithProject(project),
	}
	newOpts = append(newOpts, opts...)
	ev, err := manager.NewEvaluator(ctx, newOpts...)
	if err != nil {
		return nil, err
	}
	return &simpleEvaluator{Evaluator: ev, manager: manager}, nil
}
