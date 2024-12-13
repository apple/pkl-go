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

package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/apple/pkl-go/pkl"
)

func main() {
	runtime, err := pkl.NewExternalReaderClient(func(opts *pkl.ExternalReaderClientOptions) {
		opts.ResourceReaders = append(opts.ResourceReaders, fibReader{})
	})
	if err != nil {
		log.Fatalln(err)
	}

	if err := runtime.Run(); err != nil {
		log.Fatalln(err)
	}
}

type fibReader struct{}

var _ pkl.ResourceReader = &fibReader{}

func (r fibReader) Scheme() string {
	return "fib"
}

func (r fibReader) HasHierarchicalUris() bool {
	return false
}

func (r fibReader) IsGlobbable() bool {
	return false
}

func (r fibReader) ListElements(baseURI url.URL) ([]pkl.PathElement, error) {
	return nil, nil
}

func (r fibReader) Read(uri url.URL) ([]byte, error) {
	i, err := strconv.Atoi(uri.Opaque)
	if i <= 0 {
		err = errors.New("non-positive value")
	}
	if err != nil {
		return nil, fmt.Errorf("input uri must be in format fib:<positive integer>: %w", err)
	}

	fib := fibonacci()
	result := 0
	for range i {
		result = fib()
	}

	return []byte(strconv.Itoa(result)), nil
}

func fibonacci() func() int {
	f0, f1 := 0, 1
	return func() int {
		result := f0
		f0, f1 = f1, f0+f1
		return result
	}
}
