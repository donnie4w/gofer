// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/gosignal

package gosignal

import (
	"os"
	"os/signal"
)

func ListenSignalEvent(signalfunc func(os.Signal), sigs ...os.Signal) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)
	go func() {
		sig := <-sigChan
		signalfunc(sig)
	}()
}
