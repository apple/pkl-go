// ===----------------------------------------------------------------------===//
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
// ===----------------------------------------------------------------------===//

package pkl

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type atomicBool struct {
	atomic.Bool
}

func (a *atomicBool) get() bool {
	return a.Load()
}

func (a *atomicBool) set(value bool) {
	a.Store(value)
}

var randPool = &sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}

type atomicRandom struct{}

func (a *atomicRandom) Int63() int64 {
	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)
	return r.Int63()
}

var random = &atomicRandom{}
