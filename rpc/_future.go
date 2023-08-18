package rpc

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/basemachine/proto/pblock"
)

type Result2[A, B any] struct {
	A A
	B B
}

type Result3[A, B, C any] struct {
	A A
	B B
	C C
}

type Result4[A, B, C, D any] struct {
	A A
	B B
	C C
	D D
}

type Service interface {
	Shard(id int64) Shard
}

type Shard interface {
	Block(index int64) (Result2[pblock.Block, Stream], status.Status)

	BlockFuture(index int64) (async.Future[Result2[
		pblock.Block,
		*ref.R[Stream],
	]], status.Status)

	Block3Future(index int64) Future[Result2[
		*ref.Box[pblock.Block],
		*ref.R[Stream],
	]]

	// This one
	Block4(index int64) (*ref.Box[pblock.Block], Stream, status.Status)

	// And this one
	Block4Future(index int64) Future2[*ref.Box[pblock.Block], Stream]
}

type Future[T any] interface {
	async.Future[T]
	async.Canceller
}

type Future2[A, B any] interface {
	async.Future[Result2[A, B]]
	async.Canceller

	Results() (A, B, status.Status)
}

type Stream interface {
	Free()
}

func main3(s Service) status.Status {
	future := s.Shard(1234).Block4Future(1000)
	defer future.Cancel()

	box, stream, st := future.Results()
	if !st.OK() {
		return st
	}
	defer box.Free()
	defer stream.Free()

	return status.OK
}

func main(s Service) status.Status {
	result, st := s.Shard(123).Block(1000)
	if !st.OK() {
		return st
	}
	defer result.Free()

	_, stream := result.Unwrap()
	stream.Free()
	return status.OK
}

func main2(s Service) status.Status {
	future, st := s.Shard(123).BlockFuture(1000)
	if !st.OK() {
		return st
	}
	defer future.Cancel()

	result, st := future.Result()
	if !st.OK() {
		return st
	}
	defer result.Free()

	_, stream := result.Unwrap()
	stream.Free()
	return status.OK
}
