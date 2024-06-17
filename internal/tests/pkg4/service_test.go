package pkg4

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/internal/tests/pkg1"
	"github.com/basecomplextech/spec/rpc"
)

var _ Service = (*testService)(nil)

type testService struct{}

func newTestService() *testService {
	return &testService{}
}

func (s *testService) Subservice(ctx rpc.Context, req ServiceSubserviceRequest,
	next rpc.NextHandler[Subservice]) status.Status {
	s1 := &testSubservice{}
	return next.Handle(s1)
}

func (s *testService) Method(ctx rpc.Context) status.Status {
	return status.OK
}

func (s *testService) Method0(ctx rpc.Context, req ServiceMethod0Request) status.Status {
	return status.OK
}

func (s *testService) Method1(ctx rpc.Context, req ServiceMethod1Request) status.Status {
	return status.OK
}

func (s *testService) Method2(ctx rpc.Context, req ServiceMethod2Request) (ref.R[ServiceMethod2Response], status.Status) {
	w := NewServiceMethod2ResponseWriter()
	w.A(req.A())
	w.B(req.B())
	w.C(req.C())
	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoop(resp), status.OK
}

func (s *testService) Method3(ctx rpc.Context, req Request) (ref.R[Response], status.Status) {
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

func (s *testService) Method4(ctx rpc.Context, req ServiceMethod4Request) (ref.R[ServiceMethod4Response], status.Status) {
	w := NewServiceMethod4ResponseWriter()
	w.Ok(true)

	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoop(resp), status.OK
}

func (s *testService) Method10(ctx rpc.Context) (ref.R[ServiceMethod10Response], status.Status) {
	w := NewServiceMethod10ResponseWriter()
	w.A00(true)
	w.A01(1)
	w.A10(2)
	w.A11(3)
	w.A12(4)
	w.A20(5)
	w.A21(6)
	w.A21(7)
	w.A30(8)
	w.A31(9)
	w.A40(bin.Int64(10))
	w.A41(bin.Int128(0, 11))
	w.A42(bin.Int256(0, 0, 0, 1))

	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoop(resp), status.OK
}

func (s *testService) Method11(ctx rpc.Context) (ref.R[ServiceMethod11Response], status.Status) {
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

func (s *testService) Method20(ctx rpc.Context, ch ServiceMethod20Channel) (ref.R[ServiceMethod20Response], status.Status) {
	req, st := ch.Request()
	if !st.OK() {
		return nil, st
	}

	w := NewServiceMethod20ResponseWriter()
	w.A(req.A())
	w.B(req.B())
	w.C(req.C())

	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoop(resp), status.OK
}

func (s *testService) Method21(ctx rpc.Context, ch ServiceMethod21Channel) (ref.R[Response], status.Status) {
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

func (s *testService) Method22(ctx rpc.Context, ch ServiceMethod22Channel) (ref.R[Response], status.Status) {
	req, st := ch.Request()
	if !st.OK() {
		return nil, st
	}
	str := req.Msg().Unwrap()

	{
		w := NewOutWriter()
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

func (s *testService) Method23(ctx rpc.Context, ch ServiceMethod23Channel) (ref.R[Response], status.Status) {
	req, st := ch.Request()
	if !st.OK() {
		return nil, st
	}
	str := req.Msg().Clone()

	{
		w := NewOutWriter()
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

func (s *testSubservice) Hello(ctx rpc.Context, req SubserviceHelloRequest) (
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
