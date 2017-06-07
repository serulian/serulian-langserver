// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// WillSaveWaitUntilTextDocumentRequest defines the will-save-wait-util request, which is
// sent before a document is saved.
const WillSaveWaitUntilTextDocumentRequest = "textDocument/willSaveWaitUntil"

// WillSaveReason defines the various reasons the document will be saved.
type WillSaveReason int

const (
	// WillSaveManual indicates the save is manually triggered, e.g. by the user pressing save,
	// by starting debugging, or by an API call.
	WillSaveManual WillSaveReason = 1

	// AfterDelay indicates the save wass triggered automatically after a defined delay.
	AfterDelay = 2

	// FocusOut indicates the save was triggered due to the edito losing focus.
	FocusOut = 3
)

// WillSaveTextDocumentParams defines the parameters of the will-save-wait-util request.
type WillSaveTextDocumentParams struct {
	// TextDocument identifies the document being saved.
	TextDocument TextDocumentIdentifier `json:"textDocument"`

	// Reason is the reason this request was made.
	Reason WillSaveReason `json:"reason"`
}

// WillSaveWaitUntilTextDocumentResult defines the result of the will-save-wait-util request.
type WillSaveWaitUntilTextDocumentResult []TextEdit
