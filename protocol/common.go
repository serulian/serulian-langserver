// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package protocol mirrors the structs and methods for the Language Server protocol,
// as defined in https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md
package protocol

// TraceOption is an enumeration of the various trace operations available.
type TraceOption string

const (
	// TraceOff indicates tracing is turned off.
	TraceOff TraceOption = "off"

	// TraceMessages indicates tracing should show messages.
	TraceMessages = "messages"

	// TraceVerbose indicates verbose tracing should be enabled.
	TraceVerbose = "verbose"
)

// DocumentURI is a URI representing a document.
type DocumentURI string

func (uri DocumentURI) String() string {
	return string(uri)
}

// MarkedString represents a markdown string.
//
// MarkedString can be used to render human readable text. It is either a markdown string
// or a code-block that provides a language and a code snippet. The language identifier
// is sematically equal to the optional language identifier in fenced code blocks in GitHub
// issues. See https://help.github.com/articles/creating-and-highlighting-code-blocks/#syntax-highlighting
//
// The pair of a language and a value is an equivalent to markdown:
// ```${language}
// ${value}
// ```
//
// Note that markdown strings will be sanitized - that means html will be escaped.
type MarkedString struct {
	// Language is the markdown language.
	Language string `json:"language"`

	// Value is the markdown value.
	Value string `json:"value"`
}

// MarkupKind defines the various supported kinds of markup.
type MarkupKind string

const (
	// MarkupKindPlainText indicates that the MarkupContent is plain text.
	MarkupKindPlainText MarkupKind = "plaintext"

	// MarkupKindMarkdown indicates that the MarkupContent is markdown.
	MarkupKindMarkdown
)

// MarkupContent literal represents a string value which content is interpreted base on its
// kind flag. Currently the protocol supports `plaintext` and `markdown` as markup kinds.
//
// If the kind is `markdown` then the value can contain fenced code blocks like in GitHub issues.
// See https://help.github.com/articles/creating-and-highlighting-code-blocks/#syntax-highlighting
//
// Please Note* that clients might sanitize the return markdown. A client could decide to
// remove HTML from the markdown to avoid script execution.
//
type MarkupContent struct {
	// Kind is the type of the Markup.
	Kind MarkupKind `json:"kind"`

	// Value is the value of the Markup.
	Value string `json:"value"`
}
