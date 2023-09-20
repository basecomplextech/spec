package pkg4

import (
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

func (s *testService) Subservice(cancel <-chan struct{}, id_ bin.Bin128) (Subservice, status.Status) {
	return &testSubservice{}, status.OK
}

func (s *testService) Method(cancel <-chan struct{}) status.Status {
	return status.OK
}

func (s *testService) Method1(cancel <-chan struct{}, req ServiceMethod1Request) status.Status {
	return status.OK
}

func (s *testService) Method2(cancel <-chan struct{}, a_ int64, b_ float64, c_ bool) (
	_a int64,
	_b float64,
	_c bool,
	_st status.Status,
) {
	return a_, b_, c_, status.OK
}

func (s *testService) Method3(cancel <-chan struct{}, req ServiceMethod3Request) (
	_ok bool, _st status.Status) {
	return true, status.OK
}

func (s *testService) Method10(cancel <-chan struct{}) (
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
		bin.Bin64FromInt(10), bin.Bin128FromInt(11), bin.Bin256FromInt(1),
		status.OK
}

func (s *testService) Method11(cancel <-chan struct{}) (*ref.R[ServiceMethod11Response], status.Status) {
	w := NewServiceMethod11ResponseWriter()
	w.A50("hello")
	w.A51([]byte("world"))
	w.A60(pkg1.Enum_One)

	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}

	return ref.NewNoFreer(resp), status.OK
}

func (s *testService) Method20(cancel <-chan struct{}, ch *ServiceMethod20ServerChannel) (
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

func (s *testService) Method21(cancel <-chan struct{}, ch *ServiceMethod21ServerChannel) (*ref.R[Response], status.Status) {
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
		if st := ch.Send(cancel, msg); !st.OK() {
			return nil, st
		}
	}

	w := NewResponseWriter()
	w.Msg(str)
	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoFreer(resp), status.OK
}

func (s *testService) Method22(cancel <-chan struct{}, ch *ServiceMethod22ServerChannel) (*ref.R[Response], status.Status) {
	req, st := ch.Request()
	if !st.OK() {
		return nil, st
	}
	str := req.Msg().Unwrap()

	_, st = ch.Receive(cancel)
	if !st.OK() {
		return nil, st
	}

	w := NewResponseWriter()
	w.Msg(str)
	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoFreer(resp), status.OK
}

func (s *testService) Method23(cancel <-chan struct{}, ch *ServiceMethod23ServerChannel) (*ref.R[Response], status.Status) {
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
		if st := ch.Send(cancel, msg); !st.OK() {
			return nil, st
		}
	}

	_, st = ch.Receive(cancel)
	if !st.OK() {
		return nil, st
	}

	w := NewResponseWriter()
	w.Msg(str)
	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoFreer(resp), status.OK
}

var _ Subservice = (*testSubservice)(nil)

type testSubservice struct{}

func (s *testSubservice) Hello(cancel <-chan struct{}, req SubserviceHelloRequest) (
	*ref.R[SubserviceHelloResponse], status.Status) {
	msg := req.Msg().Clone()

	w := NewSubserviceHelloResponseWriter()
	w.Msg(msg)

	resp, err := w.Build()
	if err != nil {
		return nil, status.WrapError(err)
	}
	return ref.NewNoFreer(resp), status.OK
}
