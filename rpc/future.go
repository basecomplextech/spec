package rpc

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
)

// Future is a utility interface for an RPC future.
type Future[A any] interface {
	async.CancelFuture[A]
}

// Future2 wraps a future with two results.
type Future2[A, B any] interface {
	async.CancelFuture[Result2[A, B]]

	// Results returns the results and the status.
	Results() (A, B, status.Status)
}

// Future3 wraps a future with three results.
type Future3[A, B, C any] interface {
	async.CancelFuture[Result3[A, B, C]]

	// Results returns the results and the status.
	Results() (A, B, C, status.Status)
}

// Future4 wraps a future with four results.
type Future4[A, B, C, D any] interface {
	async.CancelFuture[Result4[A, B, C, D]]

	// Results returns the results and the status.
	Results() (A, B, C, D, status.Status)
}

// Future5 wraps a future with five results.
type Future5[A, B, C, D, E any] interface {
	async.CancelFuture[Result5[A, B, C, D, E]]

	// Results returns the results and the status.
	Results() (A, B, C, D, E, status.Status)
}

// New

// NewFuture returns a future from an async future.
func NewFuture[A any](future async.CancelFuture[A]) Future[A] {
	return future
}

// NewFuture2 returns a future from an async future.
func NewFuture2[A, B any](future async.CancelFuture[Result2[A, B]]) Future2[A, B] {
	return &future2[A, B]{future}
}

// NewFuture3 returns a future from an async future.
func NewFuture3[A, B, C any](future async.CancelFuture[Result3[A, B, C]]) Future3[A, B, C] {
	return &future3[A, B, C]{future}
}

// NewFuture4 returns a future from an async future.
func NewFuture4[A, B, C, D any](future async.CancelFuture[Result4[A, B, C, D]]) Future4[A, B, C, D] {
	return &future4[A, B, C, D]{future}
}

// NewFuture5 returns a future from an async future.
func NewFuture5[A, B, C, D, E any](future async.CancelFuture[Result5[A, B, C, D, E]]) Future5[A, B, C, D, E] {
	return &future5[A, B, C, D, E]{future}
}

// Completed

// Completed returns a resolved future.
func Completed[A any](result A, st status.Status) Future[A] {
	return async.Completed(result, st)
}

// Completed2 returns a resolved future.
func Completed2[A, B any](result Result2[A, B], st status.Status) Future2[A, B] {
	future := async.Completed(result, st)
	return NewFuture2[A, B](future)
}

// Completed3 returns a resolved future.
func Completed3[A, B, C any](result Result3[A, B, C], st status.Status) Future3[A, B, C] {
	future := async.Completed(result, st)
	return NewFuture3[A, B, C](future)
}

// Completed4 returns a resolved future.
func Completed4[A, B, C, D any](result Result4[A, B, C, D], st status.Status) Future4[A, B, C, D] {
	future := async.Completed(result, st)
	return NewFuture4[A, B, C, D](future)
}

// Completed5 returns a resolved future.
func Completed5[A, B, C, D, E any](result Result5[A, B, C, D, E], st status.Status) Future5[A, B, C, D, E] {
	future := async.Completed(result, st)
	return NewFuture5[A, B, C, D, E](future)
}

// internal

type future2[A, B any] struct {
	async.CancelFuture[Result2[A, B]]
}

type future3[A, B, C any] struct {
	async.CancelFuture[Result3[A, B, C]]
}

type future4[A, B, C, D any] struct {
	async.CancelFuture[Result4[A, B, C, D]]
}

type future5[A, B, C, D, E any] struct {
	async.CancelFuture[Result5[A, B, C, D, E]]
}

func (f *future2[A, B]) Results() (A, B, status.Status) {
	r, st := f.Result()
	return r.A, r.B, st
}

func (f *future3[A, B, C]) Results() (A, B, C, status.Status) {
	r, st := f.Result()
	return r.A, r.B, r.C, st
}

func (f *future4[A, B, C, D]) Results() (A, B, C, D, status.Status) {
	r, st := f.Result()
	return r.A, r.B, r.C, r.D, st
}

func (f *future5[A, B, C, D, E]) Results() (A, B, C, D, E, status.Status) {
	r, st := f.Result()
	return r.A, r.B, r.C, r.D, r.E, st
}
