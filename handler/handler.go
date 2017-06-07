// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package handler implements a VSCode-compatible language server for Serulian.
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/serulian/compiler/grok"

	"github.com/serulian/serulian-langserver/protocol"

	"github.com/sourcegraph/jsonrpc2"
)

type langServerState string

const (
	statePreInitialized langServerState = "pre-init"
	stateInitializing                   = "initializing"
	stateRunning                        = "running"
	stateShuttingDown                   = "shutting-down"
)

// SerulianLangServerHandler defines a JSON-RPC handler that implements the Serulian language server.
type SerulianLangServerHandler struct {
	// currentState holds the current state of the language server.
	currentState langServerState

	// rootURI holds the URI of the root for the workspace.
	rootURI protocol.DocumentURI

	// documentTracker defines a tracker for managing the state of all open documents (source files).
	documentTracker documentTracker

	// groker holds a reference to the Grok for this *workspace*, if any. The workspace Grok should
	// only be used for global lookup.
	groker *grok.Groker
}

// NewHandler creates a Serulian language server handler.
func NewHandler() jsonrpc2.Handler {
	return &SerulianLangServerHandler{
		currentState:    statePreInitialized,
		documentTracker: newDocumentTracker(),
	}
}

// Handle implements the JSON-RPC handling method for all incoming requests and notifications.
func (h *SerulianLangServerHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	// If the call must be executed synchronously, do so directly.
	if h.requiresSynchronousExecution(req.Method) {
		jsonrpc2.HandlerWithError(h.syncHandle).Handle(ctx, conn, req)
		return
	}

	// Otherwise, execute the call via a goroutine so that it doesn't block other calls.
	go jsonrpc2.HandlerWithError(h.syncHandle).Handle(ctx, conn, req)
}

// requiresSynchronousExecution returns if the given method must be executed synchronously, because it modifies
// the document tracker. Idea based on isFileSystemRequest in https://github.com/sourcegraph/go-langserver.
func (h *SerulianLangServerHandler) requiresSynchronousExecution(method string) bool {
	return method == protocol.DidOpenTextDocumentNotification ||
		method == protocol.DidChangeTextDocumentNotification ||
		method == protocol.DidCloseTextDocumentNotification
}

// syncHandle is a synchronous handler for the language server requests.
func (h *SerulianLangServerHandler) syncHandle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	log.Printf("(%v) Got request: %v\n", h.currentState, req)

	switch h.currentState {
	case statePreInitialized:
		return h.handlePreInit(ctx, conn, req)

	case stateInitializing:
		return h.handleInitializing(ctx, conn, req)

	case stateRunning:
		return h.handleRunning(ctx, conn, req)

	case stateShuttingDown:
		return h.handleShuttingDown(ctx, conn, req)
	}

	return nil, fmt.Errorf("Missing handler for method %s", req.Method)
}

// decodeParameters attempts to decode the parameters of the request into the given struct. If it fails,
// the proper JSON-RPC error is returned indicating the failure.
func (h *SerulianLangServerHandler) decodeParameters(req *jsonrpc2.Request, paramsStruct interface{}) error {
	if req.Params == nil {
		log.Printf("Missing parameters for request %s\n", req.Method)
		return &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: "Missing parameters"}
	}

	err := json.Unmarshal(*req.Params, paramsStruct)
	if err != nil {
		log.Printf("Error when parsing parameters for request %s: %v\n", req.Method, err)
		return &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: "Invalid parameters"}
	}

	return nil
}
