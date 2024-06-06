package pkg4

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/internal/tests/pkg1"
)

var _ Service = (*testService)(nil)

type testService struct{}

func newTestService() *testService {
	return &testService{}
}

func (s *testService) Subservice(ctx async.Context, id_ bin.Bin128) (Subservice, status.Status) {
	return &testSubservice{}, status.OK
}

func (s *testService) Method(ctx async.Context) status.Status {
	return status.OK
}

func (s *testService) Method1(ctx async.Context, req ServiceMethod1Request) status.Status {
	return status.OK
}

func (s *testService) Method2(ctx async.Context, a_ int64, b_ float64, c_ bool) (
	_a int64,
	_b float64,
	_c bool,
	_st status.Status,
) {
	return a_, b_, c_, status.OK
}

func (s *testService) Method3(ctx async.Context, req Request) (ref.R[Response], status.Status) {
	msg := req.Msg()

	buf := alloc.NewBuffer()
	w := NewResponseWriterBuffer(buf)
	w.Msg(msg.Unwrap())

	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewFreer(resp, buf), status.OK
}

func (s *testService) Method4(ctx async.Context, req ServiceMethod4Request) (
	_ok bool, _st status.Status) {
	return true, status.OK
}

func (s *testService) Method10(ctx async.Context) (
	_a00 bool,
	_a01 byte,
	_a10 int16,
	_a11 int32,
	_a12 int64,
	_a20 uint16,
	_a21 uint32,
	_a22 uint64,
	_a30 float32,
	_a31 float64,
	_a40 bin.Bin64,
	_a41 bin.Bin128,
	_a42 bin.Bin256,
	_st status.Status,
) {
	return true, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		bin.Int64(10), bin.Int128(0, 11), bin.Int256(0, 0, 0, 1),
		status.OK
}

func (s *testService) Method11(ctx async.Context) (ref.R[ServiceMethod11Response], status.Status) {
	w := NewServiceMethod11ResponseWriter()
	w.A50("hello")
	w.A51([]byte("world"))
	w.A60(pkg1.Enum_One)

	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}

	return ref.NewNoop(resp), status.OK
}

func (s *testService) Method20(ctx async.Context, ch ServiceMethod20Channel) (
	_a int64,
	_b float64,
	_c bool,
	_st status.Status,
) {
	a_, b_, c_, st := ch.Request()
	if !st.OK() {
		return 0, 0, false, st
	}
	return a_, b_, c_, status.OK
}

func (s *testService) Method21(ctx async.Context, ch ServiceMethod21Channel) (ref.R[Response], status.Status) {
	req, st := ch.Request()
	if !st.OK() {
		return nil, st
	}
	str := req.Msg().Unwrap()

	{
		w := NewInWriter()
		w.A(1)
		w.B(2)
		w.C("3")
		msg, err := w.Build()
		if err != nil {
		}
		if st := ch.Send(ctx, msg); !st.OK() {
			return nil, st
		}
	}

	w := NewResponseWriter()
	w.Msg(str)
	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoop(resp), status.OK
}

func (s *testService) Method22(ctx async.Context, ch ServiceMethod22Channel) (ref.R[Response], status.Status) {
	req, st := ch.Request()
	if !st.OK() {
		return nil, st
	}
	str := req.Msg().Unwrap()

	_, st = ch.Receive(ctx)
	if !st.OK() {
		return nil, st
	}

	w := NewResponseWriter()
	w.Msg(str)
	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoop(resp), status.OK
}

func (s *testService) Method23(ctx async.Context, ch ServiceMethod23Channel) (ref.R[Response], status.Status) {
	req, st := ch.Request()
	if !st.OK() {
		return nil, st
	}
	str := req.Msg().Clone()

	{
		w := NewInWriter()
		w.A(1)
		w.B(2)
		w.C("3")
		msg, err := w.Build()
		if err != nil {
		}
		if st := ch.Send(ctx, msg); !st.OK() {
			return nil, st
		}
	}

	_, st = ch.Receive(ctx)
	if !st.OK() {
		return nil, st
	}

	w := NewResponseWriter()
	w.Msg(str)
	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoop(resp), status.OK
}

var _ Subservice = (*testSubservice)(nil)

type testSubservice struct{}

func (s *testSubservice) Hello(ctx async.Context, req SubserviceHelloRequest) (
	ref.R[SubserviceHelloResponse], status.Status) {
	msg := req.Msg().Clone()

	w := NewSubserviceHelloResponseWriter()
	w.Msg(msg)

	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoop(resp), status.OK
}
