// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// DefinitionRequest defines the name of the definition method.
const DefinitionRequest = "textDocument/definition"

// DefinitionParams defines the parameters for the definition request.
type DefinitionParams struct {
	TextDocumentPositionParams
}

// DefinitionResult defines the result for the definition request.
type DefinitionResult []Location
