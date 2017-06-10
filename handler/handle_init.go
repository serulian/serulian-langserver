// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"log"
	"os"

	"github.com/serulian/compiler/grok"

	"github.com/serulian/serulian-langserver/protocol"

	"github.com/sourcegraph/jsonrpc2"
)

func (h *SerulianLangServerHandler) handlePreInit(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	// Only commands allowed in a pre-initialized state are the `initialize` command and notifications.
	if !req.Notif {
		// Make sure we have the initialize method.
		if req.Method == protocol.InitializeMethod {
			initializeParams := protocol.InitializeParams{}
			err := h.decodeParameters(req, &initializeParams)
			if err != nil {
				return nil, err
			}

			log.Printf("Got initialization request: %v\n", initializeParams)

			// Check for the parent process ID (if necessary). If the parent process was specified but is no longer
			// running, then exit.
			if initializeParams.ProcessID != nil {
				_, err := os.FindProcess(*initializeParams.ProcessID)
				if err != nil {
					log.Printf("Was given parent process ID %v but process failed to be found: %v\n", initializeParams.ProcessID, err)
					os.Exit(-1) // Parent is gone.
				}
			}

			// Initialize the document tracker.
			workspaceRoot := h.entrypointSourceFile
			if workspaceRoot == "" {
				workspaceRoot, err = h.documentTracker.uriToPath(initializeParams.RootURI.String())
				if err != nil {
					log.Printf("Error when trying to convert workspace root URI to a path: %v\n", err)
					return nil, err
				}
			}

			h.documentTracker.initializeWorkspace(ctx, conn, workspaceRoot)

			// Set the state as initializing.
			h.currentState = stateInitializing

			// Respond back with our capabilities.
			trueValue := true
			fullDoc := protocol.FullDocument
			return protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					TextDocumentSync: &protocol.TextDocumentSyncOptions{
						OpenClose:         &trueValue,
						Change:            &fullDoc,
						WillSaveWaitUntil: &trueValue,
					},
					HoverProvider:              &trueValue,
					DefinitionProvider:         &trueValue,
					WorkspaceSymbolProvider:    &trueValue,
					DocumentFormattingProvider: &trueValue,
					CompletionProvider: &protocol.CompletionOptions{
						TriggerCharacters: []string{" ", ".", "<"},
					},
					SignatureHelpProvider: &protocol.SignatureHelpOptions{
						TriggerCharacters: []string{"(", "[", ","},
					},
					CodeLensProvider: &protocol.CodeLensOptions{
						ResolveProvider: &trueValue,
					},
					ExecuteCommandProvider: &protocol.ExecuteCommandOptions{
						Commands: grok.AllActions,
					},
					CodeActionProvider: &trueValue,
				},
			}, nil
		}

		// Otherwise, we return an error with code -32002 per the spec.
		return nil, &jsonrpc2.Error{Code: -32002, Message: "Server is not initialized"}
	}

	// Only supported notification in this state is the `exit` notification.
	if req.Method == protocol.ExitNotification {
		os.Exit(-1) // -1 since we haven't received the shutdown notification.
	}

	// Ignore all other notifications.
	return nil, nil
}

func (h *SerulianLangServerHandler) handleInitializing(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	// Only events are allowed while initializing.
	if !req.Notif {
		// We return an error with code -32002 per the spec.
		return nil, &jsonrpc2.Error{Code: -32002, Message: "Server is not initialized"}
	}

	switch req.Method {
	case protocol.ExitNotification:
		os.Exit(-1) // -1 since we haven't received the shutdown notification.

	case protocol.InitializedNotification:
		h.currentState = stateRunning
		return nil, nil
	}

	// Ignore all other notifications.
	return nil, nil
}

func (h *SerulianLangServerHandler) handleShuttingDown(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	// Only supported notification in this state is the `exit` notification.
	if req.Method == protocol.ExitNotification {
		os.Exit(0) // 0 since we've received the exit notification.
	}

	// Ignore all other commands and notifications.
	return nil, nil
}
