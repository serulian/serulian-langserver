// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// DocumentFormattingRequest defines the name of the formatting method.
const DocumentFormattingRequest = "textDocument/formatting"

// DocumentFormattingParams defines the parameters for the formatting request.
type DocumentFormattingParams struct {
	// TextDocument identifies the document to format.
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentFormattingResult defines the result of a formatting request.
type DocumentFormattingResult []TextEdit
