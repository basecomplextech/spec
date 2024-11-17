package prpc

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
)

var (
	_ alloc.Buffer
	_ async.Context
	_ bin.Bin128
	_ buffer.Buffer
	_ spec.MessageTable
	_ pools.Pool[any]
	_ ref.Ref
	_ spec.Type
	_ status.Status
)

// MessageType

type MessageType int32

const (
	MessageType_Undefined MessageType = 0
	MessageType_Request   MessageType = 1
	MessageType_Response  MessageType = 2
	MessageType_Message   MessageType = 3
	MessageType_End       MessageType = 4
)

func OpenMessageType(b []byte) MessageType {
	v, _, _ := spec.DecodeInt32(b)
	return MessageType(v)
}

func DecodeMessageType(b []byte) (result MessageType, size int, err error) {
	v, size, err := spec.DecodeInt32(b)
	if err != nil || size == 0 {
		return
	}
	result = MessageType(v)
	return
}

func EncodeMessageTypeTo(b buffer.Buffer, v MessageType) (int, error) {
	return spec.EncodeInt32(b, int32(v))
}

func (e MessageType) String() string {
	switch e {
	case MessageType_Undefined:
		return "undefined"
	case MessageType_Request:
		return "request"
	case MessageType_Response:
		return "response"
	case MessageType_Message:
		return "message"
	case MessageType_End:
		return "end"
	}
	return ""
}

// Message

type Message struct {
	msg spec.Message
}

func NewMessage(msg spec.Message) Message {
	return Message{msg}
}

func OpenMessage(b []byte) Message {
	msg := spec.OpenMessage(b)
	return Message{msg}
}

func OpenMessageErr(b []byte) (_ Message, err error) {
	msg, err := spec.OpenMessageErr(b)
	if err != nil {
		return
	}
	return Message{msg}, nil
}

func ParseMessage(b []byte) (_ Message, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return Message{msg}, size, nil
}

func (m Message) Type() MessageType { return OpenMessageType(m.msg.FieldRaw(1)) }
func (m Message) Req() Request      { return NewRequest(m.msg.Message(2)) }
func (m Message) Resp() Response    { return NewResponse(m.msg.Message(3)) }
func (m Message) Msg() spec.Bytes   { return m.msg.Bytes(4) }

func (m Message) HasType() bool { return m.msg.HasField(1) }
func (m Message) HasReq() bool  { return m.msg.HasField(2) }
func (m Message) HasResp() bool { return m.msg.HasField(3) }
func (m Message) HasMsg() bool  { return m.msg.HasField(4) }

func (m Message) IsEmpty() bool                         { return m.msg.Empty() }
func (m Message) Clone() Message                        { return Message{m.msg.Clone()} }
func (m Message) CloneToArena(a alloc.Arena) Message    { return Message{m.msg.CloneToArena(a)} }
func (m Message) CloneToBuffer(b buffer.Buffer) Message { return Message{m.msg.CloneToBuffer(b)} }
func (m Message) Unwrap() spec.Message                  { return m.msg }

// Request

type Request struct {
	msg spec.Message
}

func NewRequest(msg spec.Message) Request {
	return Request{msg}
}

func OpenRequest(b []byte) Request {
	msg := spec.OpenMessage(b)
	return Request{msg}
}

func OpenRequestErr(b []byte) (_ Request, err error) {
	msg, err := spec.OpenMessageErr(b)
	if err != nil {
		return
	}
	return Request{msg}, nil
}

func ParseRequest(b []byte) (_ Request, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return Request{msg}, size, nil
}

func (m Request) Calls() spec.MessageList[Call] {
	return spec.NewMessageList(m.msg.List(1), OpenCallErr)
}
func (m Request) HasCalls() bool                        { return m.msg.HasField(1) }
func (m Request) IsEmpty() bool                         { return m.msg.Empty() }
func (m Request) Clone() Request                        { return Request{m.msg.Clone()} }
func (m Request) CloneToArena(a alloc.Arena) Request    { return Request{m.msg.CloneToArena(a)} }
func (m Request) CloneToBuffer(b buffer.Buffer) Request { return Request{m.msg.CloneToBuffer(b)} }
func (m Request) Unwrap() spec.Message                  { return m.msg }

// Call

type Call struct {
	msg spec.Message
}

func NewCall(msg spec.Message) Call {
	return Call{msg}
}

func OpenCall(b []byte) Call {
	msg := spec.OpenMessage(b)
	return Call{msg}
}

func OpenCallErr(b []byte) (_ Call, err error) {
	msg, err := spec.OpenMessageErr(b)
	if err != nil {
		return
	}
	return Call{msg}, nil
}

func ParseCall(b []byte) (_ Call, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return Call{msg}, size, nil
}

func (m Call) Method() spec.String { return m.msg.String(1) }
func (m Call) Input() spec.Message { return m.msg.Field(2).Message() }

func (m Call) HasMethod() bool { return m.msg.HasField(1) }
func (m Call) HasInput() bool  { return m.msg.HasField(2) }

func (m Call) IsEmpty() bool                      { return m.msg.Empty() }
func (m Call) Clone() Call                        { return Call{m.msg.Clone()} }
func (m Call) CloneToArena(a alloc.Arena) Call    { return Call{m.msg.CloneToArena(a)} }
func (m Call) CloneToBuffer(b buffer.Buffer) Call { return Call{m.msg.CloneToBuffer(b)} }
func (m Call) Unwrap() spec.Message               { return m.msg }

// Response

type Response struct {
	msg spec.Message
}

func NewResponse(msg spec.Message) Response {
	return Response{msg}
}

func OpenResponse(b []byte) Response {
	msg := spec.OpenMessage(b)
	return Response{msg}
}

func OpenResponseErr(b []byte) (_ Response, err error) {
	msg, err := spec.OpenMessageErr(b)
	if err != nil {
		return
	}
	return Response{msg}, nil
}

func ParseResponse(b []byte) (_ Response, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return Response{msg}, size, nil
}

func (m Response) Status() Status     { return NewStatus(m.msg.Message(1)) }
func (m Response) Result() spec.Value { return m.msg.Field(2) }

func (m Response) HasStatus() bool { return m.msg.HasField(1) }
func (m Response) HasResult() bool { return m.msg.HasField(2) }

func (m Response) IsEmpty() bool                          { return m.msg.Empty() }
func (m Response) Clone() Response                        { return Response{m.msg.Clone()} }
func (m Response) CloneToArena(a alloc.Arena) Response    { return Response{m.msg.CloneToArena(a)} }
func (m Response) CloneToBuffer(b buffer.Buffer) Response { return Response{m.msg.CloneToBuffer(b)} }
func (m Response) Unwrap() spec.Message                   { return m.msg }

// Status

type Status struct {
	msg spec.Message
}

func NewStatus(msg spec.Message) Status {
	return Status{msg}
}

func OpenStatus(b []byte) Status {
	msg := spec.OpenMessage(b)
	return Status{msg}
}

func OpenStatusErr(b []byte) (_ Status, err error) {
	msg, err := spec.OpenMessageErr(b)
	if err != nil {
		return
	}
	return Status{msg}, nil
}

func ParseStatus(b []byte) (_ Status, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return Status{msg}, size, nil
}

func (m Status) Code() spec.String    { return m.msg.String(1) }
func (m Status) Message() spec.String { return m.msg.String(2) }

func (m Status) HasCode() bool    { return m.msg.HasField(1) }
func (m Status) HasMessage() bool { return m.msg.HasField(2) }

func (m Status) IsEmpty() bool                        { return m.msg.Empty() }
func (m Status) Clone() Status                        { return Status{m.msg.Clone()} }
func (m Status) CloneToArena(a alloc.Arena) Status    { return Status{m.msg.CloneToArena(a)} }
func (m Status) CloneToBuffer(b buffer.Buffer) Status { return Status{m.msg.CloneToBuffer(b)} }
func (m Status) Unwrap() spec.Message                 { return m.msg }

// MessageWriter

type MessageWriter struct {
	w spec.MessageWriter
}

func NewMessageWriter() MessageWriter {
	w := spec.NewMessageWriter()
	return MessageWriter{w}
}

func NewMessageWriterBuffer(b buffer.Buffer) MessageWriter {
	w := spec.NewMessageWriterBuffer(b)
	return MessageWriter{w}
}

func NewMessageWriterTo(w spec.MessageWriter) MessageWriter {
	return MessageWriter{w}
}

func (w MessageWriter) Type(v MessageType) { spec.WriteField(w.w.Field(1), v, EncodeMessageTypeTo) }
func (w MessageWriter) Req() RequestWriter {
	w1 := w.w.Field(2).Message()
	return NewRequestWriterTo(w1)
}
func (w MessageWriter) CopyReq(v Request) error {
	return w.w.Field(2).Any(v.Unwrap().Raw())
}
func (w MessageWriter) Resp() ResponseWriter {
	w1 := w.w.Field(3).Message()
	return NewResponseWriterTo(w1)
}
func (w MessageWriter) CopyResp(v Response) error {
	return w.w.Field(3).Any(v.Unwrap().Raw())
}
func (w MessageWriter) Msg(v []byte) { w.w.Field(4).Bytes(v) }

func (w MessageWriter) Merge(msg Message) error {
	return w.w.Merge(msg.Unwrap())
}

func (w MessageWriter) End() error {
	return w.w.End()
}

func (w MessageWriter) Build() (_ Message, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return OpenMessageErr(bytes)
}

func (w MessageWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// RequestWriter

type RequestWriter struct {
	w spec.MessageWriter
}

func NewRequestWriter() RequestWriter {
	w := spec.NewMessageWriter()
	return RequestWriter{w}
}

func NewRequestWriterBuffer(b buffer.Buffer) RequestWriter {
	w := spec.NewMessageWriterBuffer(b)
	return RequestWriter{w}
}

func NewRequestWriterTo(w spec.MessageWriter) RequestWriter {
	return RequestWriter{w}
}

func (w RequestWriter) Calls() spec.MessageListWriter[CallWriter] {
	w1 := w.w.Field(1).List()
	return spec.NewMessageListWriter(w1, NewCallWriterTo)
}

func (w RequestWriter) Merge(msg Request) error {
	return w.w.Merge(msg.Unwrap())
}

func (w RequestWriter) End() error {
	return w.w.End()
}

func (w RequestWriter) Build() (_ Request, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return OpenRequestErr(bytes)
}

func (w RequestWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// CallWriter

type CallWriter struct {
	w spec.MessageWriter
}

func NewCallWriter() CallWriter {
	w := spec.NewMessageWriter()
	return CallWriter{w}
}

func NewCallWriterBuffer(b buffer.Buffer) CallWriter {
	w := spec.NewMessageWriterBuffer(b)
	return CallWriter{w}
}

func NewCallWriterTo(w spec.MessageWriter) CallWriter {
	return CallWriter{w}
}

func (w CallWriter) Method(v string)                { w.w.Field(1).String(v) }
func (w CallWriter) Input() spec.MessageWriter      { return w.w.Field(2).Message() }
func (w CallWriter) CopyInput(v spec.Message) error { return w.w.Field(2).Any(v.Raw()) }

func (w CallWriter) Merge(msg Call) error {
	return w.w.Merge(msg.Unwrap())
}

func (w CallWriter) End() error {
	return w.w.End()
}

func (w CallWriter) Build() (_ Call, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return OpenCallErr(bytes)
}

func (w CallWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// ResponseWriter

type ResponseWriter struct {
	w spec.MessageWriter
}

func NewResponseWriter() ResponseWriter {
	w := spec.NewMessageWriter()
	return ResponseWriter{w}
}

func NewResponseWriterBuffer(b buffer.Buffer) ResponseWriter {
	w := spec.NewMessageWriterBuffer(b)
	return ResponseWriter{w}
}

func NewResponseWriterTo(w spec.MessageWriter) ResponseWriter {
	return ResponseWriter{w}
}

func (w ResponseWriter) Status() StatusWriter {
	w1 := w.w.Field(1).Message()
	return NewStatusWriterTo(w1)
}
func (w ResponseWriter) CopyStatus(v Status) error {
	return w.w.Field(1).Any(v.Unwrap().Raw())
}
func (w ResponseWriter) Result() spec.FieldWriter      { return w.w.Field(2) }
func (w ResponseWriter) CopyResult(v spec.Value) error { return w.w.Field(2).Any(v) }

func (w ResponseWriter) Merge(msg Response) error {
	return w.w.Merge(msg.Unwrap())
}

func (w ResponseWriter) End() error {
	return w.w.End()
}

func (w ResponseWriter) Build() (_ Response, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return OpenResponseErr(bytes)
}

func (w ResponseWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// StatusWriter

type StatusWriter struct {
	w spec.MessageWriter
}

func NewStatusWriter() StatusWriter {
	w := spec.NewMessageWriter()
	return StatusWriter{w}
}

func NewStatusWriterBuffer(b buffer.Buffer) StatusWriter {
	w := spec.NewMessageWriterBuffer(b)
	return StatusWriter{w}
}

func NewStatusWriterTo(w spec.MessageWriter) StatusWriter {
	return StatusWriter{w}
}

func (w StatusWriter) Code(v string)    { w.w.Field(1).String(v) }
func (w StatusWriter) Message(v string) { w.w.Field(2).String(v) }

func (w StatusWriter) Merge(msg Status) error {
	return w.w.Merge(msg.Unwrap())
}

func (w StatusWriter) End() error {
	return w.w.End()
}

func (w StatusWriter) Build() (_ Status, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return OpenStatusErr(bytes)
}

func (w StatusWriter) Unwrap() spec.MessageWriter {
	return w.w
}
