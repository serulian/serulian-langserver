// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"sync/atomic"
	"time"
)

// debounce performs debouncing of the given function, invoking after the given interval
// has completed and no additional inputs have occurred during that time.
// Inspired by: https://nathanleclaire.com/blog/2014/08/03/write-a-function-similar-to-underscore-dot-jss-debounce-in-golang/
func debounce(f func(data interface{}), interval time.Duration) func(data interface{}) {
	var addVersion uint32

	checkAndWait := func(checkVersion uint32, data interface{}) {
		currentVersion := atomic.LoadUint32(&addVersion)
		if currentVersion != checkVersion {
			return
		}

		<-time.After(interval)
		currentVersion = atomic.LoadUint32(&addVersion)
		if currentVersion == checkVersion {
			f(data)
		}
	}

	return func(data interface{}) {
		version := atomic.AddUint32(&addVersion, 1)
		go checkAndWait(version, data)
	}
}
