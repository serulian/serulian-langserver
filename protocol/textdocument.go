// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// DidOpenTextDocumentNotification defines the name of the text-document-opened notification.
const DidOpenTextDocumentNotification = "textDocument/didOpen"

// DidChangeTextDocumentNotification defines the name of the text-document-changed notification.
const DidChangeTextDocumentNotification = "textDocument/didChange"

// DidCloseTextDocumentNotification defines the name of the text-document-closed notification.
const DidCloseTextDocumentNotification = "textDocument/didClose"

// Location represents a location in a particular document.
type Location struct {
	// URI is the URI of the document.
	URI DocumentURI `json:"uri"`

	// Range is the range in the document.
	Range Range `json:"range"`
}

// Position represents a position in a document.
type Position struct {
	// Line is the (0-indexed) line number of the position.
	Line int `json:"line"`

	// Column is the (0-indexed) column position on the line.
	Column int `json:"character"`
}

// Range represents a range in a document.
type Range struct {
	// Start is the starting position of the range, inclusive.
	Start Position `json:"start"`

	// End is the ending position of the range, exclusive.
	End Position `json:"end"`
}

// TextDocumentPositionParams defines parameters representing a position in a text document.
type TextDocumentPositionParams struct {
	// TextDocument is the text document.
	TextDocument TextDocumentIdentifier `json:"textDocument"`

	// Position is the position in the document.
	Position Position `json:"position"`
}

// TextDocumentIdentifier defines a reference to a document in the client.
type TextDocumentIdentifier struct {
	// URI is the text document's URI.
	URI DocumentURI `json:"uri"`
}

// VersionedTextDocumentIdentifier defines a reference to a specific version of a document in the client.
type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier

	// Version is a monotomically increasing version number for the document, which is increased
	// everytime the contents of the document have changed.
	Version int `json:"version"`
}

// TextDocumentItem is a single document in the client.
type TextDocumentItem struct {
	// URI is the text document's URI.
	URI DocumentURI `json:"uri"`

	// LanguageID is the ID of the language of the document, as identified by the client.
	LanguageID string `json:"languageId"`

	// Version is a monotomically increasing version number for the document, which is increased
	// everytime the contents of the document have changed.
	Version int `json:"version"`

	// Text is the contents of the opened document.
	Text string `json:"text"`
}

// DidOpenTextDocumentParams is the parameters for the DidOpenTextDocumentNotification.
type DidOpenTextDocumentParams struct {
	// TextDocument is the document that was opened.
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidChangeTextDocumentParams is the parameters for the DidChangeTextDocumentNotification.
type DidChangeTextDocumentParams struct {
	// TextDocument is the document that was changed. The version number points
	// to the version after all provided content changes have
	// been applied.
	TextDocument VersionedTextDocumentIdentifier `json:"textDocument"`

	// ContentChanges defines the changes to the document.
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// TextDocumentContentChangeEvent defines a single change event on a document.
type TextDocumentContentChangeEvent struct {
	// Text is the contents of the changed document.
	Text string `json:"text"`
}

// DidCloseTextDocumentParams is the parameters for the DidCloseTextDocumentNotification.
type DidCloseTextDocumentParams struct {
	// TextDocument is the document that was closed
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextEdit defines a single edit to a document.
type TextEdit struct {
	// Range defines the range to edit.
	Range Range `json:"range"`

	// NewText is the new text for the range.
	NewText string `json:"newText"`
}
