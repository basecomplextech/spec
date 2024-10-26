// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import "github.com/basecomplextech/baselibrary/async"

var closedChan = func() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}()

// worker pool

// workerPool allows to reuse goroutines with bigger stacks for handling channels.
// It does not provide any performance benefits in test benchmarks, but it does provide
// performance gains in real-world scenarios with big stacks, especially with chained RPC handlers.
//
//	BenchmarkTable_Get_Parallel-10:
//	goroutines		144731 ops	568 B/op	23 allocs/op
//	goroutine pool	184801 ops	558 B/op	23 allocs/op
var workerPool = async.NewRoutinePool()
