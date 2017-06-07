// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// ExecuteCommandRequest defines the name of the execute command method.
const ExecuteCommandRequest = "workspace/executeCommand"

// ExecuteCommandParams defines the parameters of the execute command request.
type ExecuteCommandParams struct {
	// Command is the internal command function name of the command to be executed.
	Command string `json:"command"`

	// Arguments are the arguments for the command.
	Arguments []interface{} `json:"arguments"`
}

// Command represents a single command sent from the server to the client, to which
// the client can respond asking the server to execute.
type Command struct {
	// Title is the title of the command to display.
	Title string `json:"title"`

	// Command is the internal command function name that will be sent to the server
	// when this command is to be executed.
	Command string `json:"command"`

	// Arguments are the arguments with which the command handler should be invoked.
	Arguments []interface{} `json:"arguments"`
}
