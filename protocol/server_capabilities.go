// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

// SaveOptions defines the options for when a document is saved in the client.
type SaveOptions struct {
	// IncludeText indicates that the client is supposed to include the content on save.
	IncludeText *bool `json:"includeText,omitempty"`
}

// TextDocumentSyncKind defines the various kinds of syncing.
type TextDocumentSyncKind int

const (
	// NoneDocument indicates that documents should not be synced at all.
	NoneDocument TextDocumentSyncKind = 0

	// FullDocument indicates that Documents are synced by always sending the full content
	// of the document.
	FullDocument TextDocumentSyncKind = 1

	// IncrementalDocument indicates that Documents are synced by sending the full content on open.
	// After that only incremental updates to the document are sent.
	IncrementalDocument TextDocumentSyncKind = 2
)

// TextDocumentSyncOptions defines the various options for syncing of documents
// between the client and server.
type TextDocumentSyncOptions struct {
	// OpenClose, if set, indicates that open and close notifications are sent to the server.
	OpenClose *bool `json:"openClose,omitempty"`

	// Change defines the change notifications are sent to the server
	Change *TextDocumentSyncKind `json:"change,omitempty"`

	// WillSave defines whether `will save` definitions are sent to the server.
	WillSave *bool `json:"willSave,omitempty"`

	// WillSaveWaitUntil defines whether saves will wait for the server to respond.
	WillSaveWaitUntil *bool `json:"willSaveWaitUntil,omitempty"`

	// Save defines the various saving options.
	Save *SaveOptions `json:"save,omitempty"`
}

// CompletionOptions defines the options for the completion feature offered by the server.
type CompletionOptions struct {
	// ResolveProvider, if true, indicates that the server provides support
	// to resolve additional information for a completion item.
	ResolveProvider *bool `json:"resolveProvider,omitempty"`

	// TriggerCharacters defines the set of characters used to trigger completion.
	TriggerCharacters []string `json:"triggerCharacters"`
}

// SignatureHelpOptions defines the options for the signature help feature offered by the server.
type SignatureHelpOptions struct {
	// TriggerCharacters defines the set of characters used to trigger signature help.
	TriggerCharacters []string `json:"triggerCharacters"`
}

// CodeLensOptions defines the options for the CodeLens feature offered by the server.
type CodeLensOptions struct {
	// ResolveProvider, if true, indicates that the server provides support
	// to resolve additional information for a code lens operation.
	ResolveProvider *bool `json:"resolveProvider,omitempty"`
}

// DocumentOnTypeFormattingOptions defines the options for the on-type formatting feature offered by the server.
type DocumentOnTypeFormattingOptions struct {
	// FirstTriggerCharacter defines a character on which formatting should be triggered, like `}`.
	FirstTriggerCharacter rune `json:"firstTriggerCharacter"`

	// MoreTriggerCharacter defines additional trigger characters for formatting.
	MoreTriggerCharacter *[]string `json:"moreTriggerCharacter,omitempty"`
}

// DocumentLinkOptions defines the various options for the document link capability.
type DocumentLinkOptions struct {
	// ResolveProvider, if true, indicates that the server provides support
	// to resolve additional information for a document link.
	ResolveProvider *bool `json:"resolveProvider,omitempty"`
}

// ExecuteCommandOptions defines the options of the various commands that can be executed
// on the server.
type ExecuteCommandOptions struct {
	// Commands defines the commands that can be executed on the server.
	Commands []string `json:"commands,omitempty"`
}

// ServerCapabilities defines the set of capabilities for the language server.
type ServerCapabilities struct {
	// TextDocumentSync defines how text documents are synced.
	TextDocumentSync *TextDocumentSyncOptions `json:"textDocumentSync,omitempty"`

	// HoverProvider indicates (if true), that hover support is provided by this server.
	HoverProvider *bool `json:"hoverProvider,omitempty"`

	// CompletionProvider indicates (if set), that this server provides completion with the
	// given options.
	CompletionProvider *CompletionOptions `json:"completionProvider,omitempty"`

	// SignatureHelpProvider indicates (if set), that this server supports signature help
	// with the given options.
	SignatureHelpProvider *SignatureHelpOptions `json:"signatureHelpProvider,omitempty"`

	// DefinitionProvider indicates (if true), that this server supports goto definitions.
	DefinitionProvider *bool `json:"definitionProvider,omitempty"`

	// ReferencesProvider indicates (if true), that this server provides find references support.
	ReferencesProvider *bool `json:"referencesProvider,omitempty"`

	// DocumentHighlightProvider indicates (if true), that this server provides document highlight support.
	DocumentHighlightProvider *bool `json:"documentHighlightProvider,omitempty"`

	// DocumentSymbolProvider indicates (if true), that this server provides document symbol support.
	DocumentSymbolProvider *bool `json:"documentSymbolProvider,omitempty"`

	// WorkspaceSymbolProvider indicates (if true), that this server provides workspace symbol support.
	WorkspaceSymbolProvider *bool `json:"workspaceSymbolProvider,omitempty"`

	// CodeActionProvider indicates (if true), that this server provides code actions.
	CodeActionProvider *bool `json:"codeActionProvider,omitempty"`

	// CodeLensProvider indicates (if set), that this server provides code lens with the given options.
	CodeLensProvider *CodeLensOptions `json:"codeLensProvider,omitempty"`

	// DocumentFormattingProvider indicates (if true), that this server provides document formatting support.
	DocumentFormattingProvider *bool `json:"documentFormattingProvider,omitempty"`

	// DocumentRangeFormattingProvider indicates (if true), that this server provides document range formatting support.
	DocumentRangeFormattingProvider *bool `json:"documentRangeFormattingProvider,omitempty"`

	// DocumentOnTypeFormattingProvider indicates (if set), that this server provides document on-type formatting support with the given options.
	DocumentOnTypeFormattingProvider *DocumentOnTypeFormattingOptions `json:"documentOnTypeFormattingProvider,omitempty"`

	// RenameProvider indicates (if true), that this server provides rename support.
	RenameProvider *bool `json:"renameProvider,omitempty"`

	// DocumentLinkProvider indicates (if set), that this server provides document link support with the given options.
	DocumentLinkProvider *DocumentLinkOptions `json:"documentLinkProvider,omitempty"`

	// ExecuteCommandProvider indicates (if set), that this server provides execute command with the given options.
	ExecuteCommandProvider *ExecuteCommandOptions `json:"executeCommandProvider,omitempty"`

	// Experimental defines all experimental features supported by this server.
	Experimental interface{} `json:"experimental,omitempty"`
}
