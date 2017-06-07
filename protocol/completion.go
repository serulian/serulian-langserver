// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// CompletionRequest defines the name of the completion method.
const CompletionRequest = "textDocument/completion"

// CompletionParams defines the parameters for the completion request.
type CompletionParams struct {
	TextDocumentPositionParams
}

// CompletionResult defines the result for the completion request.
type CompletionResult []CompletionItem

// CompletionItem represents a single completion.
type CompletionItem struct {
	// Label is the label of this completion item. By default
	// also the text that is inserted when selecting this completion.
	Label string `json:"label"`

	// Kind is the kind of this completion item.
	Kind CompletionKind `json:"kind"`

	// InsertText is a string to be inserted when this completion is selected.
	// If empty, the label is used.
	InsertText string `json:"insertText"`

	// Detail is a human-readable string with additional information
	// about this item, like type or symbol information.
	Detail string `json:"detail"`

	// Documentation is a human-readable string that represents a doc comment.
	Documentation string `json:"documentation"`
}

// CompletionKind represents the various kinds of completions.
type CompletionKind int

const (
	CompletionText        CompletionKind = 1
	CompletionMethod                     = 2
	CompletionFunction                   = 3
	CompletionConstructor                = 4
	CompletionField                      = 5
	CompletionVariable                   = 6
	CompletionClass                      = 7
	CompletionInterface                  = 8
	CompletionModule                     = 9
	CompletionProperty                   = 10
	CompletionUnit                       = 11
	CompletionValue                      = 12
	CompletionEnum                       = 13
	CompletionKeyword                    = 14
	CompletionSnippet                    = 15
	CompletionColor                      = 16
	CompletionFile                       = 17
	CompletionReference                  = 18
)
