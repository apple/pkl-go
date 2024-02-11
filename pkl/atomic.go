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
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type atomicBool struct {
	*atomic.Bool
}

func (a *atomicBool) get() bool {
	return a.Load()
}

func (a *atomicBool) set(value bool) {
	a.Store(value)
}

type atomicRandom struct {
	mutex sync.Mutex
	rand  *rand.Rand
}

func (a *atomicRandom) Int63() int64 {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.rand.Int63()
}

var random = &atomicRandom{
	rand: rand.New(rand.NewSource(time.Now().UnixMilli())),
}
