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
)

type Class struct {
	HasMembers
	Name     string
	Extends  string
	Abstract bool
	model    *Model
}

func NewClass(m *Model, name string) *Class {
	return &Class{
		HasMembers: HasMembers{members: map[string]*Member{}},
		Name:       name,
		model:      m,
	}
}

func (m *Class) WritePkl(buf io.StringWriter, indent string) {
	buf.WriteString(indent)
	if m.Abstract {
		buf.WriteString("abstract ")
	}
	buf.WriteString("class ")
	buf.WriteString(m.Name)

	if m.Extends != "" {
		buf.WriteString(" extends ")
		buf.WriteString(m.Extends)
	}
	buf.WriteString(" {\n")

	m.HasMembers.WritePkl(buf, indent+DefaultIndent)

	buf.WriteString("}\n\n")
}
