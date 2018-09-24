// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// WorkspaceSymbolRequest defines the name of the workspace symbol method.
const WorkspaceSymbolRequest = "workspace/symbol"

// WorkspaceSymbolParams defines the parameters for the workspace symbol request.
type WorkspaceSymbolParams struct {
	// Query is the lookup query.
	Query string `json:"query"`
}

// WorkspaceSymbolResponse defines the response to the workspace symbol request.
type WorkspaceSymbolResponse []SymbolInformation

// SymbolKind is an enumeration of the different kinds of symbols.
type SymbolKind int

const (
	SymbolFile          SymbolKind = 1
	SymbolModule                   = 2
	SymbolNamespace                = 3
	SymbolPackage                  = 4
	SymbolClass                    = 5
	SymbolMethod                   = 6
	SymbolProperty                 = 7
	SymbolField                    = 8
	SymbolConstructor              = 9
	SymbolEnum                     = 10
	SymbolInterface                = 11
	SymbolFunction                 = 12
	SymbolVariable                 = 13
	SymbolConstant                 = 14
	SymbolString                   = 15
	SymbolNumber                   = 16
	SymbolBoolean                  = 17
	SymbolArray                    = 18
	SymbolObject                   = 19
	SymbolKey                      = 20
	SymbolNull                     = 21
	SymbolEnumMember               = 22
	SymbolStruct                   = 23
	SymbolEvent                    = 24
	SymbolOperator                 = 25
	SymbolTypeParameter            = 26
)

// SymbolInformation represents a single symbol found in a document or workspace.
type SymbolInformation struct {
	// Name is the name of the symbol.
	Name string `json:"name"`

	// Kind is the kind of the symbol.
	Kind SymbolKind `json:"kind"`

	// Location is the location of the symbol in source.
	Location Location `json:"location"`

	// ContainerName is the name of the symbol containing this symbol, if any.
	ContainerName *string `json:"containerName,omitempty"`
}
