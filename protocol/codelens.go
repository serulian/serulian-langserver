// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// CodeLensRequest defines the name of the code lens lookup method.
const CodeLensRequest = "textDocument/codeLens"

// CodeLensParams defines the parameters for the codelens request.
type CodeLensParams struct {
	// TextDocument is the document for which CodeLens is being requested.
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// CodeLensResult defines the result of the CodeLens lookup request.
type CodeLensResult []CodeLens

// ResolveCodeLensRequest defines the name of the resolve-code lens method.
const ResolveCodeLensRequest = "codeLens/resolve"

// ResolveCodeLensParams defines the parameters for the resolve code lens request.
type ResolveCodeLensParams CodeLens

// ResolveCodeLensResult defines the result for the resolve code lens request.
type ResolveCodeLensResult CodeLens

// CodeLens represents a single CodeLens in a source document.
type CodeLens struct {
	// Range is the range in the document to which the CodeLens applies. Should not apply to multiple lines.
	Range Range `json:"range"`

	// Command is the command to display for this CodeLens. Should be omitted on the initial response and only
	// added when the resolve is called.
	Command *Command `json:"command,omitempty"`

	// Data is a data entry field that is preserved on a code lens item between
	// a code lens and a code lens resolve request.
	Data interface{} `json:"data,omitempty"`
}
