// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// CodeActionRequest defines the name of the code actions lookup method.
const CodeActionRequest = "textDocument/codeAction"

// CodeActionParams are the parameters for the code actions.
type CodeActionParams struct {
	// TextDocument is the document for which the code actions are being requested.
	TextDocument TextDocumentIdentifier `json:"textDocument"`

	// Range is the range for which the code actions are being requested.
	Range Range `json:"range"`
}

// CodeActionResult represents the result of a code actions lookup request.
type CodeActionResult []Command
