// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// PublicDiagonsticsNotification defines a notification from the *server* to the client
// about diagnostic information being available for a document.
const PublicDiagonsticsNotification = "textDocument/publishDiagnostics"

// PublishDiagnosticsParams defines the parameters for the PublicDiagonsticsNotification.
type PublishDiagnosticsParams struct {
	// URI is the URI of the document for which we are publishing diagnostics.
	URI DocumentURI `json:"uri"`

	// Diagnostics is the set of diagnostic information being published.
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Diagnostic defines a single piece of diagnostic information about a document.
type Diagnostic struct {
	// Range defines the range in the document to which this information applies.
	Range Range `json:"range"`

	// Severity defines the severity of this information.
	Severity DiagnosticSeverity `json:"severity"`

	// Code defines an optional code for this information.
	Code *string `json:"code"`

	// Source defines a human-readable string describing the source of this
	// diagnostic, e.g. 'typescript' or 'super lint'.
	Source *string `json:"source"`

	// Messages defines the human-readable message for this information.
	Message string `json:"message"`
}

// DiagnosticSeverity defines the various severity levels for diagnostic information.
type DiagnosticSeverity int

const (
	// DiagnosticError indicates an error-level severity.
	DiagnosticError DiagnosticSeverity = 1

	// DiagnosticWarning indicates an warning-level severity.
	DiagnosticWarning = 2

	// DiagnosticInformation indicates an informative-level severity.
	DiagnosticInformation = 3

	// DiagnosticHint indicates a hint-level severity.
	DiagnosticHint = 4
)
