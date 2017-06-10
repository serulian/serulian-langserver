// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"log"

	"github.com/serulian/compiler/compilercommon"

	"github.com/serulian/serulian-langserver/protocol"

	"github.com/sourcegraph/jsonrpc2"
)

type diagnoseParams struct {
	dt                  *documentTracker
	path                string
	version             int
	isWorkspaceDiagnose bool

	ctx  context.Context
	conn *jsonrpc2.Conn
}

// diagnoseDocument is a helper which is invoked as a goroutine to background index and report diagnostics for
// a document.
func diagnoseDocument(data interface{}) {
	params := data.(diagnoseParams)
	dt := params.dt
	ctx := params.ctx
	conn := params.conn
	path := params.path
	version := params.version
	isWorkspaceDiagnose := params.isWorkspaceDiagnose
	workspaceGroker := dt.workspaceGrok

	log.Printf("Starting diagnoseDocument for %s at version %v (isWorkspace=%v) \n", path, version, isWorkspaceDiagnose)

	// If this is a diagnose for a non-workspace document, then make sure we are still at the correct version.
	groker := workspaceGroker
	pathsToReport := []string{path}

	if isWorkspaceDiagnose {
		pathsToReport = dt.documents.Keys()
	} else {
		// Ensure we are still at the current version.
		current, valid := dt.getDocumentAtVersion(path, version)
		if !valid {
			log.Printf("Canceled (#1) diagnoseDocument for %s at version %v", path, version)
			return
		}

		// Retrieve the handle.
		groker = current.groker
		if groker == nil {
			log.Printf("No groker for diagnoseDocument for %s at version %v", path, version)
			return
		}
	}

	handle, err := groker.GetHandle()
	if err != nil {
		log.Printf("Encountered error retrieving handle for diagnoseDocument for %s at version %v", path, version)
		return
	}

	log.Printf("Got handle with status %v for diagnoseDocument for %s at version %v", handle.IsCompilable(), path, version)

	// Ensure we are still at the current version.
	if !isWorkspaceDiagnose {
		_, valid := dt.getDocumentAtVersion(path, version)
		if !valid {
			log.Printf("Canceled (#2) diagnoseDocument for %s at version %v", path, version)
			return
		}
	}

	// Collect any issues found, by document.
	for _, currentPath := range pathsToReport {
		// Skip any paths no longer referenced by the document. They'll be updated on next edit.
		if !handle.ContainsSource(compilercommon.InputSource(currentPath)) {
			continue
		}

		var issues = []protocol.Diagnostic{}
		addIssue := func(sourceRange compilercommon.SourceRange, message string, severity protocol.DiagnosticSeverity) {
			documentRange, err := dt.convertRange(sourceRange)
			if err != nil {
				return
			}

			issues = append(issues, protocol.Diagnostic{
				Severity: severity,
				Message:  message,
				Range:    documentRange,
			})
		}

		for _, sourceError := range handle.Errors() {
			if string(sourceError.SourceRange().Source()) == currentPath {
				addIssue(sourceError.SourceRange(), sourceError.Error(), protocol.DiagnosticError)
			}
		}

		for _, sourceWarning := range handle.Warnings() {
			if string(sourceWarning.SourceRange().Source()) == currentPath {
				addIssue(sourceWarning.SourceRange(), sourceWarning.Warning(), protocol.DiagnosticWarning)
			}
		}

		uri, okay := dt.sourceToURI(compilercommon.InputSource(currentPath))
		if !okay {
			log.Printf("Could not convert path `%s` to URI in diagnoseDocument at version %v", currentPath, version)
			continue
		}

		if !isWorkspaceDiagnose {
			// Ensure we are still at the current version.
			_, valid := dt.getDocumentAtVersion(currentPath, version)
			if !valid {
				continue
			}
		}

		err = conn.Notify(ctx, protocol.PublicDiagonsticsNotification, protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: issues,
		})

		if err != nil {
			log.Printf("Notify failed for diagnoseDocument for %s at version %v: %v", currentPath, version, err)
			continue
		}
	}
}
