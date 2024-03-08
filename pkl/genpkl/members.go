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
package genpkl

import (
	"io"
	"reflect"
	"sync"
)

type Member struct {
	Name string
	Type string
}

type HasMembers struct {
	members map[string]*Member
	lock    sync.Mutex
}

func NewMember(name string, typeName string) *Member {
	return &Member{
		Name: name,
		Type: typeName,
	}
}

func (m *HasMembers) GetOrCreateField(name, typeName string, field reflect.StructField) *Member {
	m.lock.Lock()
	defer m.lock.Unlock()

	// TODO track which types are used from which module so we can handle multi-module ownership of types better?

	if m.members == nil {
		m.members = map[string]*Member{}
	}
	answer := m.members[name]
	if answer == nil {
		answer = NewMember(name, typeName)
		m.members[name] = answer
	}
	return answer
}

func (m *HasMembers) WritePkl(buf io.StringWriter, indent string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, m := range m.members {
		buf.WriteString(indent)
		buf.WriteString(m.Name)
		buf.WriteString(": ")
		buf.WriteString(m.Type)
		buf.WriteString("\n")
	}
}
