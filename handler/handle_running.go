// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"log"
	"os"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/compilerutil"
	"github.com/serulian/compiler/graphs/typegraph"
	"github.com/serulian/compiler/grok"

	"github.com/serulian/serulian-langserver/protocol"

	"strings"

	"github.com/sourcegraph/jsonrpc2"
)

func (h *SerulianLangServerHandler) handleRunning(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request, cancelationHandle *CancelationHandle) (result interface{}, err error) {
	switch req.Method {
	// Code Actions.
	case protocol.CodeActionRequest:
		params := protocol.CodeActionParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		log.Printf("Got code actions request for document %v with range: %v\n", params.TextDocument.URI, params.Range)
		if !h.documentTracker.isTracking(params.TextDocument.URI.String()) {
			log.Printf("Not tracking document %s\n", params.TextDocument.URI)
			return protocol.CodeActionResult([]protocol.Command{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Grab a Grok handle for the document.
		uri := params.TextDocument.URI
		handle, document, err := h.documentTracker.getGrokHandleAndDocument(uri.String(), grok.HandleAllowStale)
		if err != nil {
			log.Printf("Got error when trying to get grok handle for %s: %v", uri, err)
			return protocol.CodeActionResult([]protocol.Command{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Retrieve the path.
		path, err := h.documentTracker.uriToPath(uri.String())
		if err != nil {
			log.Printf("Got error when trying to convert URI to path for %s: %v", uri, err)
			return protocol.CodeActionResult([]protocol.Command{}), nil
		}

		// Retrieve the actions.
		source := compilercommon.InputSource(path)
		actions, err := handle.GetActionsForPosition(source, params.Range.Start.Line, params.Range.Start.Column)
		if err != nil {
			log.Printf("Got error when trying to retrieve actions for path %s: %v", uri, err)
			return protocol.CodeActionResult([]protocol.Command{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Return the actions, converted to commands.
		commands := make([]protocol.Command, 0, len(actions))
		for _, action := range actions {
			commands = append(commands, protocol.Command{
				Title:     action.Title,
				Command:   string(action.Action),
				Arguments: []interface{}{path, document.version, action.ActionParams},
			})
		}

		return protocol.CodeActionResult(commands), nil

	// Execute command.
	case protocol.ExecuteCommandRequest:
		params := protocol.ExecuteCommandParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		log.Printf("Got execute command request for command %s with arguments: %v\n", params.Command, params.Arguments)

		// Read the document path and version from the arguments.
		path := params.Arguments[0].(string)
		version := int(params.Arguments[1].(float64))

		// Find the document for that version, if any.
		document, found := h.documentTracker.getDocumentAtVersion(path, version)
		if !found {
			return nil, nil
		}

		// Execute the action via the document's groker, if any.
		groker := document.groker
		if groker == nil {
			return nil, nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		handle, err := groker.GetHandleWithOption(grok.HandleAllowStale)
		if err != nil {
			log.Printf("Got error when trying to get grok handle for %s: %v\n", path, err)
			return nil, nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		source := compilercommon.InputSource(path)
		err = handle.ExecuteAction(grok.Action(params.Command), params.Arguments[2].(map[string]interface{}), source)
		if err != nil {
			log.Printf("Got error when trying to execute command %s with arguments %v: %v\n", params.Command, params.Arguments, err)
			return nil, nil
		}

		return nil, nil

	// CodeLens.
	case protocol.CodeLensRequest:
		params := protocol.CodeLensParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		log.Printf("CodeLens request for document %s", params.TextDocument.URI)
		if !h.documentTracker.isTracking(params.TextDocument.URI.String()) {
			log.Printf("Not tracking document %s\n", params.TextDocument.URI)
			return protocol.CodeLensResult([]protocol.CodeLens{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Grab a Grok handle for the document.
		uri := params.TextDocument.URI
		handle, document, err := h.documentTracker.getGrokHandleAndDocument(uri.String(), grok.HandleAllowStale)
		if err != nil {
			log.Printf("Got error when trying to get grok handle for %s: %v", uri, err)
			return protocol.CodeLensResult([]protocol.CodeLens{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Retrieve the path.
		path, err := h.documentTracker.uriToPath(uri.String())
		if err != nil {
			log.Printf("Got error when trying to convert URI to path for %s: %v", uri, err)
			return protocol.CodeLensResult([]protocol.CodeLens{}), nil
		}

		// Lookup the CodeContextAndAction's for the path.
		source := compilercommon.InputSource(path)
		ccas, err := handle.GetContextActions(source)
		if err != nil {
			log.Printf("Got error when trying to lookup CCA's for %s: %v", uri, err)
			return protocol.CodeLensResult([]protocol.CodeLens{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// For each CCA found, save it to the document's map under a unique ID and return it to the client.
		codeLenses := make([]protocol.CodeLens, 0, len(ccas))
		for _, cca := range ccas {
			convertedRange, err := h.documentTracker.convertRange(cca.Range)
			if err != nil {
				continue
			}

			ccaID := compilerutil.NewUniqueId()
			document.codeContextOrActions.Set(ccaID, cca)

			codeLenses = append(codeLenses, protocol.CodeLens{
				Range: convertedRange,
				Data: map[string]interface{}{
					"id":      ccaID,
					"path":    path,
					"version": document.version,
				},
			})
		}

		return protocol.CodeLensResult(codeLenses), nil

	// Resolve CodeLens.
	case protocol.ResolveCodeLensRequest:
		params := protocol.ResolveCodeLensParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		dataMap := params.Data.(map[string]interface{})
		ccaID := dataMap["id"].(string)
		path := dataMap["path"].(string)
		version := int(dataMap["version"].(float64))

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Find the CodeContextOrAction in the tracker.
		document, found := h.documentTracker.getDocumentAtVersion(path, version)
		if !found {
			return protocol.ResolveCodeLensResult(protocol.CodeLens(params)), nil
		}

		foundCCA, found := document.codeContextOrActions.Get(ccaID)
		if !found {
			return protocol.ResolveCodeLensResult(protocol.CodeLens(params)), nil
		}

		cca := foundCCA.(grok.CodeContextOrAction)
		contextOrAction, ok := cca.Resolve()
		if !ok {
			return protocol.ResolveCodeLensResult(protocol.CodeLens(params)), nil
		}

		return protocol.ResolveCodeLensResult(
			protocol.CodeLens{
				Range: params.Range,
				Command: &protocol.Command{
					Title:     contextOrAction.Title,
					Command:   string(contextOrAction.Action),
					Arguments: []interface{}{path, version, contextOrAction.ActionParams},
				},
				Data: params.Data,
			}), nil

	// Signature help request.
	case protocol.SignatureHelpRequest:
		params := protocol.SignatureHelpParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		log.Printf("Signature help request for document %s", params.TextDocument.URI)
		if !h.documentTracker.isTracking(params.TextDocument.URI.String()) {
			log.Printf("Not tracking document %s\n", params.TextDocument.URI)
			return protocol.SignatureHelpResult{[]protocol.SignatureInformation{}, 0, 0}, nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Grab a Grok handle for the document.
		uri := params.TextDocument.URI
		handle, err := h.documentTracker.getGrokHandle(uri.String(), grok.HandleAllowStale)
		if err != nil {
			log.Printf("Got error when trying to get grok handle for %s: %v", uri, err)
			return protocol.SignatureHelpResult{[]protocol.SignatureInformation{}, 0, 0}, nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Retrieve the path.
		path, err := h.documentTracker.uriToPath(uri.String())
		if err != nil {
			log.Printf("Got error when trying to convert URI to path for %s: %v", uri, err)
			return protocol.SignatureHelpResult{[]protocol.SignatureInformation{}, 0, 0}, nil
		}

		// Find the position's line text in the document.
		lineText, err := h.documentTracker.getLineText(uri.String(), params.Position.Line, params.Position.Column)
		if err != nil {
			log.Printf("Got error when trying to retrieve line text for %s: %v", uri, err)
			return protocol.SignatureHelpResult{[]protocol.SignatureInformation{}, 0, 0}, nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		log.Printf("Retrieving signature help for file %s and text: %s", uri, lineText)

		// Lookup the signature help via Grok.
		source := compilercommon.InputSource(path)
		signatureInformation, err := handle.GetSignatureForPosition(strings.TrimSpace(lineText), source, params.Position.Line, params.Position.Column)
		if err != nil {
			log.Printf("Got error when retrieving signature for %s: %v", uri, err)
			return protocol.SignatureHelpResult{[]protocol.SignatureInformation{}, 0, 0}, nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		if signatureInformation.Name == "" && len(signatureInformation.Parameters) == 0 {
			// TODO: can we do this in a better way?
			// Try again on a full handle.
			handle, err := h.documentTracker.getGrokHandle(uri.String(), grok.HandleMustBeFresh)
			if err != nil {
				log.Printf("Got error when trying to get fresh grok handle for %s: %v", uri, err)
				return protocol.SignatureHelpResult{[]protocol.SignatureInformation{}, 0, 0}, nil
			}

			if cancelationHandle.WasCanceled() {
				return nil, cancelationHandle.Error()
			}

			signatureInformation, err = handle.GetSignatureForPosition(strings.TrimSpace(lineText), source, params.Position.Line, params.Position.Column)
			if err != nil {
				log.Printf("Got error when signature completions for %s: %v", uri, err)
				return protocol.SignatureHelpResult{[]protocol.SignatureInformation{}, 0, 0}, nil
			}
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		if signatureInformation.Name == "" && len(signatureInformation.Parameters) == 0 {
			log.Printf("No signature found for %s: %v", uri, err)
			return protocol.SignatureHelpResult{[]protocol.SignatureInformation{}, 0, 0}, nil
		}

		fullSignature := signatureInformation.Name
		if signatureInformation.Member != nil && signatureInformation.Member.IsOperator() {
			fullSignature += "["
		} else {
			fullSignature += "("
		}

		parameters := make([]protocol.ParameterInformation, len(signatureInformation.Parameters))
		for index, parameterInfo := range signatureInformation.Parameters {
			label := ""
			if parameterInfo.Name != "" {
				label = parameterInfo.Name
			}

			if !parameterInfo.TypeReference.IsVoid() {
				if label != "" {
					label = label + " "
				}

				label = label + parameterInfo.TypeReference.String()
			}

			if index > 0 {
				fullSignature += ", "
			}
			fullSignature += label

			parameters[index] = protocol.ParameterInformation{
				Label:         label,
				Documentation: markdownContent(parameterInfo.Documentation),
			}
		}

		if signatureInformation.Member != nil && signatureInformation.Member.IsOperator() {
			fullSignature += "]"
		} else {
			fullSignature += ")"
		}

		return protocol.SignatureHelpResult{
			Signatures: []protocol.SignatureInformation{
				protocol.SignatureInformation{
					Label:         fullSignature,
					Documentation: markdownContent(signatureInformation.Documentation),
					Parameters:    parameters,
				},
			},
			ActiveSignatureIndex: 0,
			ActiveParameterIndex: signatureInformation.ActiveParameterIndex,
		}, nil

	// Completion request.
	case protocol.CompletionRequest:
		params := protocol.CompletionParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		log.Printf("Completion request for document %s", params.TextDocument.URI)
		if !h.documentTracker.isTracking(params.TextDocument.URI.String()) {
			log.Printf("Not tracking document %s\n", params.TextDocument.URI)
			return protocol.CompletionResult([]protocol.CompletionItem{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Grab a Grok handle for the document.
		uri := params.TextDocument.URI
		handle, err := h.documentTracker.getGrokHandle(uri.String(), grok.HandleAllowStale)
		if err != nil {
			log.Printf("Got error when trying to get grok handle for %s: %v", uri, err)
			return protocol.CompletionResult([]protocol.CompletionItem{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Retrieve the path.
		path, err := h.documentTracker.uriToPath(uri.String())
		if err != nil {
			log.Printf("Got error when trying to convert URI to path for %s: %v", uri, err)
			return protocol.CompletionResult([]protocol.CompletionItem{}), nil
		}

		// Find the position's line text in the document.
		lineText, err := h.documentTracker.getLineText(uri.String(), params.Position.Line, params.Position.Column)
		if err != nil {
			log.Printf("Got error when trying to retrieve line text for %s: %v", uri, err)
			return protocol.CompletionResult([]protocol.CompletionItem{}), nil
		}

		log.Printf("Retrieving completions for file %s and text: `%s`", uri, lineText)

		// Lookup the completion via Grok.
		source := compilercommon.InputSource(path)
		completionInfo, err := handle.GetCompletionsForPosition(strings.TrimSpace(lineText), source, params.Position.Line, params.Position.Column)
		if err != nil {
			log.Printf("Got error when retrieving completions for %s: %v", uri, err)
			return protocol.CompletionResult([]protocol.CompletionItem{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		if len(completionInfo.Completions) == 0 {
			// TODO: can we do this in a better way?
			// Try again on a full handle.
			handle, err := h.documentTracker.getGrokHandle(uri.String(), grok.HandleMustBeFresh)
			if err != nil {
				log.Printf("Got error when trying to get fresh grok handle for %s: %v", uri, err)
				return protocol.CompletionResult([]protocol.CompletionItem{}), nil
			}

			if cancelationHandle.WasCanceled() {
				return nil, cancelationHandle.Error()
			}

			completionInfo, err = handle.GetCompletionsForPosition(strings.TrimSpace(lineText), source, params.Position.Line, params.Position.Column)
			if err != nil {
				log.Printf("Got error when retrieving completions for %s: %v", uri, err)
				return protocol.CompletionResult([]protocol.CompletionItem{}), nil
			}
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		if len(completionInfo.Completions) == 0 {
			log.Printf("No completions found for %s: %v", uri, err)
			return protocol.CompletionResult([]protocol.CompletionItem{}), nil
		}

		completionItems := make([]protocol.CompletionItem, 0, len(completionInfo.Completions))
		for _, completionInfo := range completionInfo.Completions {
			completionKind := protocol.CompletionText
			details := ""

			switch completionInfo.Kind {
			case grok.SnippetCompletion:
				completionKind = protocol.CompletionSnippet

			case grok.TypeCompletion:
				switch completionInfo.Type.TypeKind() {
				case typegraph.GenericType:
					completionKind = protocol.CompletionTypeParameter

				case typegraph.StructType:
					completionKind = protocol.CompletionStruct

				default:
					completionKind = protocol.CompletionClass
				}

			case grok.MemberCompletion:
				if completionInfo.Member != nil {
					_, isConstructor := completionInfo.Member.ConstructorType()
					_, hasReturnType := completionInfo.Member.ReturnType()

					switch {
					case completionInfo.Member.IsOperator():
						completionKind = protocol.CompletionOperator

					case isConstructor:
						completionKind = protocol.CompletionConstructor

					case completionInfo.Member.IsField():
						completionKind = protocol.CompletionField

					case hasReturnType:
						completionKind = protocol.CompletionFunction

					default:
						completionKind = protocol.CompletionProperty
					}
				} else {
					completionKind = protocol.CompletionProperty
				}

			case grok.ImportCompletion:
				completionKind = protocol.CompletionFile

			case grok.ValueCompletion:
				completionKind = protocol.CompletionValue

			case grok.ParameterCompletion:
				completionKind = protocol.CompletionValue

			case grok.VariableCompletion:
				completionKind = protocol.CompletionVariable

			default:
				panic("Unknown kind of completion")
			}

			if !completionInfo.TypeReference.IsVoid() {
				details = completionInfo.TypeReference.String()
			}

			completionItems = append(completionItems, protocol.CompletionItem{
				Label:         completionInfo.Title,
				InsertText:    completionInfo.Code,
				Kind:          completionKind,
				Detail:        details,
				Documentation: markdownContent(completionInfo.Documentation),
			})
		}

		return protocol.CompletionResult(completionItems), nil

	// Document will save request.
	case protocol.WillSaveWaitUntilTextDocumentRequest:
		params := protocol.WillSaveTextDocumentParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		if params.Reason != protocol.WillSaveManual {
			log.Printf("Save is not a manual action; skipping formatting\n")
			return protocol.WillSaveWaitUntilTextDocumentResult([]protocol.TextEdit{}), nil
		}

		log.Printf("Document %s is about to be saved", params.TextDocument.URI)
		if !h.documentTracker.isTracking(params.TextDocument.URI.String()) {
			log.Printf("Not tracking document %s\n", params.TextDocument.URI)
			return protocol.WillSaveWaitUntilTextDocumentResult([]protocol.TextEdit{}), nil
		}

		// Format the document via the tracker.
		edits := h.documentTracker.formatDocument(string(params.TextDocument.URI))
		return protocol.WillSaveWaitUntilTextDocumentResult(edits), nil

	// Document formatting request.
	case protocol.DocumentFormattingRequest:
		params := protocol.DocumentFormattingParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		log.Printf("Got formatting request for document: %s", params.TextDocument.URI)

		// Format the document via the tracker.
		edits := h.documentTracker.formatDocument(string(params.TextDocument.URI))
		return protocol.DocumentFormattingResult(edits), nil

	// Workspace symbol lookup.
	case protocol.WorkspaceSymbolRequest:
		params := protocol.WorkspaceSymbolParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		log.Printf("Got workspace symbol request with query: %s", params.Query)
		groker := h.documentTracker.workspaceGrok
		if groker == nil {
			log.Printf("No workspace Grok available\n")
			return protocol.WorkspaceSymbolResponse([]protocol.SymbolInformation{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Grab a Grok handle.
		handle, err := groker.GetHandleWithOption(grok.HandleAllowStale)
		if err != nil {
			log.Printf("Got error when trying to get grok handle for global workspace: %v", err)
			return protocol.WorkspaceSymbolResponse([]protocol.SymbolInformation{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Perform symbol lookup.
		symbols, err := handle.FindSymbols(params.Query)
		if err != nil {
			log.Printf("Got error when trying to find symbol %s in global workspace: %v", params.Query, err)
			return protocol.WorkspaceSymbolResponse([]protocol.SymbolInformation{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		// Convert the symbols and return.
		symbolInfo := make([]protocol.SymbolInformation, 0, len(symbols))
		for _, symbol := range symbols {
			if converted, success := h.symbolInfoFromSymbol(symbol); success {
				symbolInfo = append(symbolInfo, converted)
			}
		}
		return protocol.WorkspaceSymbolResponse(symbolInfo), nil

	// Definition.
	case protocol.DefinitionRequest:
		params := protocol.DefinitionParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		log.Printf("Got definition request for document %s at position %v:%v\n", params.TextDocument.URI, params.Position.Line, params.Position.Column)
		rangeInfo, _, status := h.lookupRange(params.TextDocument.URI, params.Position, cancelationHandle)
		if !status {
			log.Printf("No valid range found for document %s at position %v:%v\n", params.TextDocument.URI, params.Position.Line, params.Position.Column)
			return protocol.DefinitionResult([]protocol.Location{}), nil
		}

		if cancelationHandle.WasCanceled() {
			return nil, cancelationHandle.Error()
		}

		locations := h.documentTracker.convertRanges(rangeInfo.SourceRanges)
		return protocol.DefinitionResult(locations), nil

	// Hover.
	case protocol.HoverRequest:
		params := protocol.HoverParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		log.Printf("Got hover request for document %s at position %v:%v\n", params.TextDocument.URI, params.Position.Line, params.Position.Column)
		rangeInfo, _, status := h.lookupRange(params.TextDocument.URI, params.Position, cancelationHandle)
		if !status {
			return protocol.HoverResult{
				Contents: []interface{}{},
			}, nil
		}

		markedTexts := rangeInfo.HumanReadable()
		markedStrings := make([]interface{}, len(markedTexts))
		for index, hr := range markedTexts {
			if hr.Kind == grok.SerulianCodeText {
				markedStrings[index] = protocol.MarkedString{"serulian", hr.Value}
			} else {
				markedStrings[index] = hr.Value
			}
		}

		return protocol.HoverResult{
			Contents: markedStrings,
		}, nil

	// Document opened.
	case protocol.DidOpenTextDocumentNotification:
		params := protocol.DidOpenTextDocumentParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		if h.documentTracker.tracksLanguage(params.TextDocument.LanguageID) {
			log.Printf("Document opened: %s\n", params.TextDocument.URI)
			h.documentTracker.openDocument(ctx, conn, params.TextDocument.URI.String(), params.TextDocument.Text, params.TextDocument.Version)
		}
		return nil, nil

	// Document updated.
	case protocol.DidChangeTextDocumentNotification:
		params := protocol.DidChangeTextDocumentParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		log.Printf("Document updated: %s\n", params.TextDocument.URI)
		h.documentTracker.updateDocument(ctx, conn, params.TextDocument.URI.String(), params.ContentChanges[0].Text, params.TextDocument.Version)
		return nil, nil

	// Document closed.
	case protocol.DidCloseTextDocumentNotification:
		params := protocol.DidCloseTextDocumentParams{}
		err := h.decodeParameters(req, &params)
		if err != nil {
			return nil, err
		}

		if h.documentTracker.tracksLanguage(params.TextDocument.LanguageID) {
			log.Printf("Document closed: %s\n", params.TextDocument.URI)
			h.documentTracker.closeDocument(params.TextDocument.URI.String())
		}
		return nil, nil

	// Exit notification.
	case protocol.ExitNotification:
		os.Exit(-1) // -1 since we haven't received the shutdown notification.
	}

	return nil, nil
}

// lookupRange performs lookup of the range matching the given position in the given document.
func (h *SerulianLangServerHandler) lookupRange(uri protocol.DocumentURI, position protocol.Position, cancelationHandle *CancelationHandle) (grok.RangeInformation, error, bool) {
	// Make sure we are tracking this document.
	if !h.documentTracker.isTracking(uri.String()) {
		log.Printf("Not tracking document %s\n", uri)
		return grok.RangeInformation{}, nil, false
	}

	if cancelationHandle.WasCanceled() {
		return grok.RangeInformation{}, cancelationHandle.Error(), false
	}

	// Grab a Grok handle for the document.
	handle, err := h.documentTracker.getGrokHandle(uri.String(), grok.HandleMustBeFresh)
	if err != nil {
		log.Printf("Got error when trying to get grok handle for %s: %v", uri, err)
		return grok.RangeInformation{}, err, false
	}

	if cancelationHandle.WasCanceled() {
		return grok.RangeInformation{}, cancelationHandle.Error(), false
	}

	// Retrieve the path.
	path, err := h.documentTracker.uriToPath(uri.String())
	if err != nil {
		log.Printf("Got error when trying to convert URI to path for %s: %v", uri, err)
		return grok.RangeInformation{}, err, false
	}

	// Lookup the position via Grok.
	source := compilercommon.InputSource(path)
	rangeInfo, err := handle.LookupPosition(source, position.Line, position.Column)
	if err != nil {
		log.Printf("Got error when trying to lookup range for %s: %v", uri, err)
		return grok.RangeInformation{}, err, false
	}

	// If we found a result, turn it into code and report it.
	log.Printf("Got result: %v => %v", rangeInfo.Kind, rangeInfo.HumanReadable())
	return rangeInfo, nil, rangeInfo.Kind != grok.NotFound
}

// symbolInfoFromSymbol converts a Grok Symbol into a SymbolInformation struct.
func (h *SerulianLangServerHandler) symbolInfoFromSymbol(symbol grok.Symbol) (protocol.SymbolInformation, bool) {
	if len(symbol.SourceRanges) < 1 {
		return protocol.SymbolInformation{}, false
	}

	var containerName *string
	var symbolKind = protocol.SymbolFile

	switch symbol.Kind {
	case grok.TypeSymbol:
		switch symbol.Type.TypeKind() {
		case typegraph.StructType:
			symbolKind = protocol.SymbolStruct

		case typegraph.AgentType:
			fallthrough

		case typegraph.ClassType:
			symbolKind = protocol.SymbolClass

		case typegraph.NominalType:
			symbolKind = protocol.SymbolObject

		case typegraph.ImplicitInterfaceType:
			fallthrough

		case typegraph.ExternalInternalType:
			symbolKind = protocol.SymbolInterface

		case typegraph.AliasType:
			fallthrough

		case typegraph.GenericType:
			symbolKind = protocol.SymbolTypeParameter

		default:
			panic("Unknown kind of type")
		}

	case grok.MemberSymbol:
		_, isConstructor := symbol.Member.ConstructorType()
		_, hasReturnType := symbol.Member.ReturnType()

		if containingType, hasContainingType := symbol.Member.ParentType(); hasContainingType {
			typeName := containingType.Name()
			containerName = &typeName
		}

		switch {
		case symbol.Member.IsOperator():
			symbolKind = protocol.SymbolOperator

		case isConstructor:
			symbolKind = protocol.SymbolConstructor

		case symbol.Member.IsField():
			symbolKind = protocol.SymbolField

		case hasReturnType:
			symbolKind = protocol.SymbolFunction

		default:
			symbolKind = protocol.SymbolProperty
		}

	case grok.ModuleSymbol:
		symbolKind = protocol.SymbolFile

	default:
		panic("Unknown kind of Grok symbol")
	}

	ranges := h.documentTracker.convertRanges(symbol.SourceRanges)
	if len(ranges) == 0 {
		return protocol.SymbolInformation{}, false
	}

	return protocol.SymbolInformation{
		Name:          symbol.Name,
		Kind:          symbolKind,
		ContainerName: containerName,
		Location:      ranges[0],
	}, true
}

// markdownContent returns the value wrapped into a MarkupContent indicating it is markdown.
func markdownContent(value string) protocol.MarkupContent {
	return protocol.MarkupContent{
		Kind:  protocol.MarkupKindMarkdown,
		Value: value,
	}
}
