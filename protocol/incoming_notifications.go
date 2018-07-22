// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import (
	"github.com/sourcegraph/jsonrpc2"
)

// ExitNotification defines the name of the `exit` notification.
const ExitNotification = "exit"

// InitializedNotification defines the name of the `initialized` notification.
const InitializedNotification = "initialized"

// CancelRequestNotification defines the name of the `cancel request` notification.
const CancelRequestNotification = "$/cancelRequest"

// CancelRequestParams are the parameters for the cancelation of a request.
type CancelRequestParams struct {
	ID jsonrpc2.ID `json:"id"`
}
