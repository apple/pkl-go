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
	"net/url"

	"github.com/apple/pkl-go/pkl/internal/msgapi"
)

// Reader is the base implementation shared by a ResourceReader and a ModuleReader.
type Reader interface {
	// Scheme returns the scheme part of the URL that this reader can read.
	// The value should be the URI scheme up to (not including) ":"
	Scheme() string

	// IsGlobbable tells if this reader supports globbing via Pkl's `import*` and `glob*` keywords
	IsGlobbable() bool

	// HasHierarchicalUris tells if the URIs handled by this reader are hierarchical.
	// Hierarchical URIs are URIs that have hierarchy elements like host, origin, query, and
	// fragment.
	//
	// A hierarchical URI must start with a "/" in its scheme specific part. For example, consider
	// the following two URIS:
	//
	//   flintstone:/persons/fred.pkl
	//   flintstone:persons/fred.pkl
	//
	// The first URI conveys name "fred.pkl" within parent "/persons/". The second URI
	// conveys the name "persons/fred.pkl" with no hierarchical meaning.
	HasHierarchicalUris() bool

	// ListElements returns the list of elements at a specified path.
	// If HasHierarchicalUris is false, path will be empty and ListElements should return all
	// available values.
	//
	// This method is only called if it is hierarchical and local, or if it is globbable.
	ListElements(url url.URL) ([]PathElement, error)
}

// PathElement is an element within a base URI.
//
// For example, a PathElement with name "bar.txt" and is not a directory at base URI "file:///foo/"
// implies URI resource `file:///foo/bar.txt`.
type PathElement interface {
	// Name is the name of the path element.
	Name() string

	// IsDirectory tells if the path element is a directory.
	IsDirectory() bool
}

type pathElement struct {
	name string

	isDirectory bool
}

func (elem *pathElement) Name() string {
	return elem.name
}

func (elem *pathElement) IsDirectory() bool {
	return elem.isDirectory
}

// NewPathElement returns an instance of PathElement.
func NewPathElement(name string, isDirectory bool) PathElement {
	return &pathElement{name: name, isDirectory: isDirectory}
}

// ResourceReader is a custom resource reader for Pkl.
//
// A ResourceReader registers the scheme that it is responsible for reading via Reader.Scheme. For
// example, a resource reader can declare that it reads a resource at secrets:MY_SECRET by returning
// "secrets" when Reader.Scheme is called.
//
// Resources are cached by Pkl for the lifetime of an Evaluator. Therefore, cacheing is not needed
// on the Go side as long as the same Evaluator is used.
//
// Resources are read via the following Pkl expressions:
//
//		 read("myscheme:myresourcee")
//		 read?("myscheme:myresource")
//	  read*("myscheme:pattern*") // only if the resource is globabble
//
// To provide a custom reader, register it on EvaluatorOptions.ResourceReaders when building
// an Evaluator.
type ResourceReader interface {
	Reader

	// Read reads the byte contents of this resource.
	Read(url url.URL) ([]byte, error)
}

// ModuleReader is a custom module reader for Pkl.
//
// A ModuleReader registers the scheme that it is responsible for reading via Reader.Scheme. For
// example, a module reader can declare that it reads a resource at myscheme:myFile.pkl by returning
// "myscheme" when Reader.Scheme is called.
//
// Modules are cached by Pkl for the lifetime of an Evaluator. Therefore, cacheing is not needed
// on the Go side as long as the same Evaluator is used.
//
// Modules are read in Pkl via the import declaration:
//
//		import "myscheme:/myFile.pkl"
//	 import* "myscheme:/*.pkl" // only when the reader is globbable
//
// Or via the import expression:
//
//		import("myscheme:myFile.pkl")
//	 import*("myscheme:/myFile.pkl") // only when the reader is globbable
//
// To provide a custom reader, register it on EvaluatorOptions.ModuleReaders when building
// an Evaluator.
type ModuleReader interface {
	Reader

	// IsLocal tells if the resources represented by this reader is considered local to the runtime.
	// A local module reader enables resolving triple-dot imports.
	IsLocal() bool

	// Read reads the string contents of this module.
	Read(url url.URL) (string, error)
}

func resourceReadersToMessage(readers []ResourceReader) []*msgapi.ResourceReader {
	resourceReaders := make([]*msgapi.ResourceReader, len(readers))
	for idx, reader := range readers {
		resourceReaders[idx] = &msgapi.ResourceReader{
			Scheme:              reader.Scheme(),
			IsGlobbable:         reader.IsGlobbable(),
			HasHierarchicalUris: reader.HasHierarchicalUris(),
		}
	}
	return resourceReaders
}

func moduleReadersToMessage(readers []ModuleReader) []*msgapi.ModuleReader {
	moduleReaders := make([]*msgapi.ModuleReader, len(readers))
	for idx, reader := range readers {
		moduleReaders[idx] = &msgapi.ModuleReader{
			Scheme:              reader.Scheme(),
			IsGlobbable:         reader.IsGlobbable(),
			HasHierarchicalUris: reader.HasHierarchicalUris(),
			IsLocal:             reader.IsLocal(),
		}
	}
	return moduleReaders
}

func externalReadersToMessage(readers map[string]ExternalReader) map[string]*msgapi.ExternalReader {
	externalReaders := make(map[string]*msgapi.ExternalReader, len(readers))
	for scheme, reader := range readers {
		externalReaders[scheme] = reader.toMessage()
	}
	return externalReaders
}
