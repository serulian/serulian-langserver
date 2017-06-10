// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/serulian/compiler/builder"
	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/formatter"
	"github.com/serulian/compiler/grok"
	"github.com/serulian/compiler/packageloader"

	"github.com/serulian/serulian-langserver/protocol"

	"path"

	"github.com/sourcegraph/jsonrpc2"
	cmap "github.com/streamrail/concurrent-map"
)

// DiagnoseDelay is the delay waited before a document is re-parsed.
const DiagnoseDelay = 500 * time.Millisecond

// getPackageLibraries returns the libraries to load, if any, when creating a Groker.
func getPackageLibraries(path string) []packageloader.Library {
	return []packageloader.Library{builder.CORE_LIBRARY}
}

// document represents a single document found in the tracker.
type document struct {
	// path is the local file system path of the document.
	path string

	// contents are the current contents of the document.
	contents string

	// version is the current version of the document.
	version int

	// groker holds a reference to the Grok for this document.
	groker *grok.Groker

	// codeContextOrActions holds the map of CodeContextOrActions's for this document, by ID.
	codeContextOrActions cmap.ConcurrentMap
}

// documentTracker keeps track of all documents (source files) which are open
// in the client.
type documentTracker struct {
	// documents is the map of documents being tracked, keyed by path.
	documents cmap.ConcurrentMap

	// localPathLoader is an instance of a LocalFilePathLoader to be used as a basis
	// for the document tracker's path loading for files that are *not* being tracked.
	localPathLoader packageloader.LocalFilePathLoader

	// debouncedDiagnose is a debounced-wrapped call over the diagnoseDocument function.
	debouncedDiagnose func(data interface{})

	// vcsDevelopmentDirectories are the specified VCS development directories to be passed
	// to Grok, if any.
	vcsDevelopmentDirectories []string

	// workspaceRootPath is the root path of the workspace.
	workspaceRootPath string

	// workspaceGrok is (if defined) the workspace-wide Grok.
	workspaceGrok *grok.Groker
}

func newDocumentTracker(vcsDevelopmentDirectories []string) *documentTracker {
	return &documentTracker{
		documents:         cmap.New(),
		localPathLoader:   packageloader.LocalFilePathLoader{},
		debouncedDiagnose: debounce(diagnoseDocument, DiagnoseDelay),

		vcsDevelopmentDirectories: vcsDevelopmentDirectories,

		workspaceRootPath: "",
		workspaceGrok:     nil,
	}
}

// initializeWorkspace initializes the document tracker over the given workspace root.
func (dt *documentTracker) initializeWorkspace(workspaceRootPath string) {
	dt.workspaceRootPath = workspaceRootPath
	dt.workspaceGrok = grok.NewGrokerWithPathLoader(workspaceRootPath, dt.vcsDevelopmentDirectories, getPackageLibraries(workspaceRootPath), dt)
}

// tracksLanguage returns true if the given language is tracked by the document tracker.
func (dt *documentTracker) tracksLanguage(languageID string) bool {
	return languageID == "serulian" || languageID == "webidl"
}

// uriToPath converts the given URI into a local file system path. If the URI is not a `file:///`
// URI, returns an error.
func (dt *documentTracker) uriToPath(uri string) (string, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	if url.Scheme != "file" {
		return "", fmt.Errorf("Can only work on local files, found: %s", uri)
	}

	return url.Path, nil
}

// isTracking returns true if and only if the document with the specified URI is already being tracked.
func (dt *documentTracker) isTracking(uri string) bool {
	path, err := dt.uriToPath(uri)
	if err != nil {
		return false
	}

	return dt.documents.Has(path)
}

// openDocument starts tracking of a document with the given URI, contents and initial version number.
func (dt *documentTracker) openDocument(ctx context.Context, conn *jsonrpc2.Conn, uri string, contents string, version int) {
	path, err := dt.uriToPath(uri)
	if err != nil {
		return
	}

	groker := grok.NewGrokerWithPathLoader(path, dt.vcsDevelopmentDirectories, getPackageLibraries(path), dt)
	dt.documents.Set(path, document{
		path:                 path,
		contents:             contents,
		version:              version,
		groker:               groker,
		codeContextOrActions: cmap.New(),
	})

	dt.debouncedDiagnose(diagnoseParams{dt, path, version, ctx, conn})
}

// updateDocument updates the contents of the document with the given URI. If the document is not being
// tracked, does nothing.
func (dt *documentTracker) updateDocument(ctx context.Context, conn *jsonrpc2.Conn, uri string, contents string, version int) {
	path, err := dt.uriToPath(uri)
	if err != nil {
		return
	}

	if !dt.documents.Has(path) {
		return
	}

	dt.documents.Upsert(path, nil, func(exists bool, valueInMap interface{}, newValue interface{}) interface{} {
		if !exists {
			return document{
				path:                 path,
				contents:             contents,
				version:              version,
				codeContextOrActions: cmap.New(),
			}
		}

		return document{
			path:                 path,
			contents:             contents,
			version:              version,
			groker:               valueInMap.(document).groker,
			codeContextOrActions: valueInMap.(document).codeContextOrActions,
		}
	})

	dt.debouncedDiagnose(diagnoseParams{dt, path, version, ctx, conn})
}

// closeDocument stops tracking the document with the given URI.
func (dt *documentTracker) closeDocument(uri string) {
	path, err := dt.uriToPath(uri)
	if err != nil {
		return
	}

	dt.documents.Remove(path)
}

// getDocumentAtVersion returns the document at the specified version, for the specified path, if any.
func (dt *documentTracker) getDocumentAtVersion(path string, version int) (document, bool) {
	currentValue, exists := dt.documents.Get(path)
	if !exists {
		return document{}, false
	}

	current := currentValue.(document)
	return current, current.version == version
}

// getGrokHandle returns the Grok handle using the given URI as the root source path.
func (dt *documentTracker) getGrokHandle(uri string, freshnessOption grok.HandleFreshnessOption) (grok.Handle, error) {
	handle, _, err := dt.getGrokHandleAndDocument(uri, freshnessOption)
	return handle, err
}

// getGrokHandleAndDocument returns the Grok handle using the given URI as the root source path.
func (dt *documentTracker) getGrokHandleAndDocument(uri string, freshnessOption grok.HandleFreshnessOption) (grok.Handle, document, error) {
	path, err := dt.uriToPath(uri)
	if err != nil {
		return grok.Handle{}, document{}, err
	}

	current, exists := dt.documents.Get(path)
	if !exists {
		return grok.Handle{}, document{}, fmt.Errorf("Document is not being tracked: %s", uri)
	}

	groker := current.(document).groker
	if groker == nil {
		return grok.Handle{}, current.(document), fmt.Errorf("Document is not being tracked: %s", uri)
	}

	handle, err := groker.GetHandleWithOption(freshnessOption)
	return handle, current.(document), err
}

// sourceToURI returns the given source as a URI.
func (dt *documentTracker) sourceToURI(source compilercommon.InputSource) (protocol.DocumentURI, bool) {
	url := url.URL{
		Scheme: "file",
		Path:   string(source),
	}

	return protocol.DocumentURI(url.String()), true
}

// convertRange returns the given source range as a Document Range.
func (dt *documentTracker) convertRange(sourceRange compilercommon.SourceRange) (protocol.Range, error) {
	startLine, startCol, err := sourceRange.Start().LineAndColumn()
	if err != nil {
		return protocol.Range{}, err
	}

	endLine, endCol, err := sourceRange.End().LineAndColumn()
	if err != nil {
		return protocol.Range{}, err
	}

	return protocol.Range{
		Start: protocol.Position{startLine, startCol},
		End:   protocol.Position{endLine, endCol},
	}, nil
}

// convertRanges converts the given SourceRanges's into Document Locations.
func (dt *documentTracker) convertRanges(sourceRanges []compilercommon.SourceRange) []protocol.Location {
	locations := make([]protocol.Location, 0, len(sourceRanges))
	for _, sourceRange := range sourceRanges {
		uri, ok := dt.sourceToURI(sourceRange.Source())
		if !ok {
			continue
		}

		documentRange, err := dt.convertRange(sourceRange)
		if err != nil {
			continue
		}

		locations = append(locations, protocol.Location{
			URI:   uri,
			Range: documentRange,
		})
	}
	return locations
}

// getLineText returns the text found on the line of the given position before its column position.
func (dt *documentTracker) getLineText(uri string, lineNumber int, colPosition int) (string, error) {
	path, err := dt.uriToPath(uri)
	if err != nil {
		return "", err
	}

	currentValue, exists := dt.documents.Get(path)
	if !exists {
		return "", fmt.Errorf("Missing document for path %s", path)
	}

	lines := strings.Split(currentValue.(document).contents, "\n")
	if lineNumber >= len(lines) {
		return "", fmt.Errorf("Invalid line %v for path %s", lineNumber, path)
	}

	lineText := lines[lineNumber]
	if colPosition > len(lineText) {
		return "", fmt.Errorf("Invalid column %v for path %s", colPosition, path)
	}

	return lineText[0:colPosition], nil
}

// formatDocument formats the document found at the given URI, return a set of edits.
func (dt *documentTracker) formatDocument(uri string) []protocol.TextEdit {
	path, err := dt.uriToPath(uri)
	if err != nil {
		return []protocol.TextEdit{}
	}

	currentValue, exists := dt.documents.Get(path)
	if !exists {
		return []protocol.TextEdit{}
	}

	current := currentValue.(document)
	if len(current.contents) == 0 {
		return []protocol.TextEdit{}
	}

	formatted, err := formatter.FormatSource(current.contents)
	if err != nil {
		log.Printf("Error when trying to format source for URI %s: %v", uri, err)
		return []protocol.TextEdit{}
	}

	// Skip if nothing has changed.
	if formatted == current.contents {
		return []protocol.TextEdit{}
	}

	lines := strings.Split(current.contents, "\n")
	lastLine := len(lines) - 1
	lastLineLength := len(lines[len(lines)-1])

	changeAll := protocol.TextEdit{
		NewText: formatted,
		Range: protocol.Range{
			protocol.Position{0, 0},
			protocol.Position{lastLine, lastLineLength},
		},
	}
	return []protocol.TextEdit{changeAll}
}

func (dt *documentTracker) VCSPackageDirectory(entrypoint packageloader.Entrypoint) string {
	workspaceRootDirectory := dt.workspaceRootPath
	if dt.IsSourceFile(workspaceRootDirectory) {
		workspaceRootDirectory = path.Dir(workspaceRootDirectory)
	}

	workspacePackageDirectory := path.Join(workspaceRootDirectory, packageloader.SerulianPackageDirectory)
	return workspacePackageDirectory
}

func (dt *documentTracker) LoadSourceFile(path string) ([]byte, error) {
	currentValue, exists := dt.documents.Get(path)
	if exists {
		return []byte(currentValue.(document).contents), nil
	}

	return dt.localPathLoader.LoadSourceFile(path)
}

func (dt *documentTracker) GetRevisionID(path string) (int64, error) {
	currentValue, exists := dt.documents.Get(path)
	if exists {
		return int64(currentValue.(document).version), nil
	}

	return dt.localPathLoader.GetRevisionID(path)
}

func (dt *documentTracker) IsSourceFile(path string) bool {
	return dt.localPathLoader.IsSourceFile(path)
}

func (dt *documentTracker) LoadDirectory(path string) ([]packageloader.DirectoryEntry, error) {
	return dt.localPathLoader.LoadDirectory(path)
}
