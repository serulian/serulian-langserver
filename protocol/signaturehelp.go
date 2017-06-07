// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// SignatureHelpRequest defines the name of the signature help method.
const SignatureHelpRequest = "textDocument/signatureHelp"

// SignatureHelpParams defines the parameters for the signature help request.
type SignatureHelpParams struct {
	TextDocumentPositionParams
}

// SignatureHelpResult defines the result for the signature help request.
type SignatureHelpResult struct {
	// Signatures defines the signatures returned, if any.
	Signatures []SignatureInformation `json:"signatures"`

	// ActiveSignatureIndex defines which signature specified is active.
	ActiveSignatureIndex int `json:"activeSignature"`

	// ActiveParameterIndex is the index in the active signature of the active parameter.
	ActiveParameterIndex int `json:"activeParameter"`
}

// SignatureInformation is information representing the signature of a single function or operator
// that is being called.
type SignatureInformation struct {
	// Label is the label to display for this signature. Typically the name and maybe the
	// type of the function/operator.
	Label string `json:"label"`

	// Documentation is the documentation to display, if any.
	Documentation string `json:"documentation"`

	// Parameters are the parameters for this signature.
	Parameters []ParameterInformation `json:"parameters"`
}

// ParameterInformation represents information about a parameter of a function or operator.
type ParameterInformation struct {
	// Label is the label to display for this parameter. Typically the name and maybe the
	// type of the function/operator.
	Label string `json:"label"`

	// Documentation is the documentation to display, if any.
	Documentation string `json:"documentation"`
}
