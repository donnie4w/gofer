// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/buffer

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
