// Copyright 2024 Ivan Korobkov. All rights reserved.

package mpx

var closedChan = func() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}()
