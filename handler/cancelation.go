// Copyright 2018 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"log"

	"github.com/sourcegraph/jsonrpc2"
)

// CancelationHandle is a handle for tracking the cancelation of a language
// server request.
type CancelationHandle struct {
	id          jsonrpc2.ID
	wasCanceled bool
}

// NewCancelationHandle returns a new cancelation handle.
func NewCancelationHandle(id jsonrpc2.ID) *CancelationHandle {
	return &CancelationHandle{id: id, wasCanceled: false}
}

// Cancel marks the operation as having been canceled.
func (ch *CancelationHandle) Cancel() {
	log.Printf("Request %s was canceled", ch.id.String())
	ch.wasCanceled = true
}

// WasCanceled returns whether the operation was canceled.
func (ch *CancelationHandle) WasCanceled() bool {
	return ch.wasCanceled
}

func (ch *CancelationHandle) Error() error {
	return &jsonrpc2.Error{
		Code:    -32800,
		Message: "Request was canceled",
	}
}
