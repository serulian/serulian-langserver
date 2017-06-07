// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// HoverRequest defines the name of the hover method.
const HoverRequest = "textDocument/hover"

// HoverParams defines the parameters for the hover request.
type HoverParams struct {
	TextDocumentPositionParams
}

// HoverResult defines the result for the hover request.
type HoverResult struct {
	// Contents is the contents to display for the hover.
	Contents []interface{} `json:"contents,omitempty"`

	// Range, if specified, indicates the highlighting range for the hover.
	Range *Range `json:"range,omitempty"`
}
