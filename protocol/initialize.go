// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// InitializeMethod defines the name of the `initialize` method.
const InitializeMethod = "initialize"

// InitializeParams is the set of parameters sent by the client for the `initialize` call.
type InitializeParams struct {
	/**
	 * The process Id of the parent process that started
	 * the server. Is null if the process has not been started by another process.
	 * If the parent process is not alive then the server should exit (see exit notification) its process.
	 */
	ProcessID *int `json:"processId,omitempty"`

	/**
	 * The rootUri of the workspace. Is null if no
	 * folder is open.
	 */
	RootURI DocumentURI `json:"rootUri,omitempty"`

	/**
	 * The initial trace setting. If omitted trace is disabled ('off').
	 */
	Trace *TraceOption `json:"trace,omitempty"`
}

// InitializeResult is the server result for an `initialize`.
type InitializeResult struct {
	/**
	 * The capabilities the language server provides.
	 */
	Capabilities ServerCapabilities `json:"capabilities"`
}
