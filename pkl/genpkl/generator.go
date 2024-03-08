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
	"fmt"
	"github.com/apple/pkl-go/pkl"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
)

const (
	// DefaultDirWritePermissions default permissions when creating a directory
	DefaultDirWritePermissions = 0o766

	// DefaultFileWritePermissions default permissions when creating a file
	DefaultFileWritePermissions = 0o644

	// DefaultIndent default pkl indentation
	DefaultIndent = "  "
)

type Generator struct {
	Modules []interface{}

	OutDir string

	// Namer allows customizing of type names. The default is to use the type's name
	// provided by the `reflect` package.
	Namer func(reflect.Type) string

	// AlternativeStructTags allows reusing existing tags on a struct field if it has no 'pkl' tag
	// e.g. reuse 'json' or 'proto'
	AlternativeStructTags []string

	model Model
}

func (o *Generator) GeneratePkl() error {
	if o.OutDir == "" {
		o.OutDir = "."
	}
	for i := 0; i < 2; i++ {
		for _, m := range o.Modules {
			err := o.GeneratePklModuleFromType(reflect.TypeOf(m), i)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *Generator) GeneratePklModuleFromType(t reflect.Type, pass int) error {
	o.model.generator = o
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("request should be a struct but was %s", t.String())
	}

	fname := o.typeName(t)

	mod := o.model.GetModule(fname, true)

	if pass == 0 {
		// first pass just load all the modules so we can separate modules from classes
		return nil
	}
	fmt.Printf("generating Module %s\n", fname)

	numFields := t.NumField()
	for i := 0; i < numFields; i++ {
		field := t.Field(i)

		// TODO handle anonymous
		if field.Name == "" {
			continue
		}

		// ignore lower case fields
		firstChar := []rune(field.Name)[0]
		if !unicode.IsUpper(firstChar) {
			continue
		}

		name := o.getFieldName(&field)
		ft := field.Type
		typeName := o.typeName(ft)
		mod.GetOrCreateField(name, typeName, field)

		mod.GetOrCreateClass(typeName, ft)
	}

	path := filepath.Join(o.OutDir, fname+".pkl")
	return mod.SavePkl(path)
}

func (o *Generator) getFieldName(field *reflect.StructField) string {
	tagValue, exists := field.Tag.Lookup(pkl.StructTag)
	if exists {
		return tagValue
	}
	for _, tag := range o.AlternativeStructTags {
		tagValue, exists = field.Tag.Lookup(tag)
		if exists {
			idx := strings.Index(tagValue, ",")
			if idx > 0 {
				tagValue = tagValue[:idx]
			}
			return tagValue
		}
	}
	return ""
}

func (o *Generator) typeName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		return o.typeName(t) + "?"
	}
	if o.Namer != nil {
		if name := o.Namer(t); name != "" {
			return name
		}
	}
	return Capitalize(t.Name())
}
