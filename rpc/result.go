package rpc

type Result[A any] struct {
	A A
}

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

type Result5[A, B, C, D, E any] struct {
	A A
	B B
	C C
	D D
	E E
}

func (r Result[A]) Unwrap() A                            { return r.A }
func (r Result2[A, B]) Unwrap() (A, B)                   { return r.A, r.B }
func (r Result3[A, B, C]) Unwrap() (A, B, C)             { return r.A, r.B, r.C }
func (r Result4[A, B, C, D]) Unwrap() (A, B, C, D)       { return r.A, r.B, r.C, r.D }
func (r Result5[A, B, C, D, E]) Unwrap() (A, B, C, D, E) { return r.A, r.B, r.C, r.D, r.E }
