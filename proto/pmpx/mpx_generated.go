package pmpx

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/encoding"
)

var (
	_ alloc.Buffer
	_ async.Context
	_ bin.Bin128
	_ buffer.Buffer
	_ encoding.MessageMeta
	_ pools.Pool[any]
	_ ref.Ref
	_ spec.Type
	_ status.Status
)

// Version

type Version int32

const (
	Version_Undefined Version = 0
	Version_Version10 Version = 10
)

func NewVersion(b []byte) Version {
	v, _, _ := encoding.DecodeInt32(b)
	return Version(v)
}

func ParseVersion(b []byte) (result Version, size int, err error) {
	v, size, err := encoding.DecodeInt32(b)
	if err != nil || size == 0 {
		return
	}
	result = Version(v)
	return
}

func WriteVersion(b buffer.Buffer, v Version) (int, error) {
	return encoding.EncodeInt32(b, int32(v))
}

func (e Version) String() string {
	switch e {
	case Version_Undefined:
		return "undefined"
	case Version_Version10:
		return "version_1_0"
	}
	return ""
}

// Code

type Code int32

const (
	Code_Undefined       Code = 0
	Code_ConnectRequest  Code = 1
	Code_ConnectResponse Code = 2
	Code_Batch           Code = 3
	Code_ChannelOpen     Code = 10
	Code_ChannelClose    Code = 11
	Code_ChannelData     Code = 12
	Code_ChannelWindow   Code = 13
)

func NewCode(b []byte) Code {
	v, _, _ := encoding.DecodeInt32(b)
	return Code(v)
}

func ParseCode(b []byte) (result Code, size int, err error) {
	v, size, err := encoding.DecodeInt32(b)
	if err != nil || size == 0 {
		return
	}
	result = Code(v)
	return
}

func WriteCode(b buffer.Buffer, v Code) (int, error) {
	return encoding.EncodeInt32(b, int32(v))
}

func (e Code) String() string {
	switch e {
	case Code_Undefined:
		return "undefined"
	case Code_ConnectRequest:
		return "connect_request"
	case Code_ConnectResponse:
		return "connect_response"
	case Code_Batch:
		return "batch"
	case Code_ChannelOpen:
		return "channel_open"
	case Code_ChannelClose:
		return "channel_close"
	case Code_ChannelData:
		return "channel_data"
	case Code_ChannelWindow:
		return "channel_window"
	}
	return ""
}

// Message

type Message struct {
	msg spec.Message
}

func NewMessage(b []byte) Message {
	msg := spec.NewMessage(b)
	return Message{msg}
}

func NewMessageErr(b []byte) (_ Message, err error) {
	msg, err := spec.NewMessageErr(b)
	if err != nil {
		return
	}
	return Message{msg}, nil
}

func MakeMessage(msg spec.Message) Message {
	return Message{msg}
}

func ParseMessage(b []byte) (_ Message, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return Message{msg}, size, nil
}

func (m Message) Code() Code                       { return NewCode(m.msg.FieldRaw(1)) }
func (m Message) ConnectRequest() ConnectRequest   { return MakeConnectRequest(m.msg.Message(2)) }
func (m Message) ConnectResponse() ConnectResponse { return MakeConnectResponse(m.msg.Message(3)) }
func (m Message) Batch() Batch                     { return MakeBatch(m.msg.Message(4)) }
func (m Message) ChannelOpen() ChannelOpen         { return MakeChannelOpen(m.msg.Message(10)) }
func (m Message) ChannelClose() ChannelClose       { return MakeChannelClose(m.msg.Message(11)) }
func (m Message) ChannelData() ChannelData         { return MakeChannelData(m.msg.Message(12)) }
func (m Message) ChannelWindow() ChannelWindow     { return MakeChannelWindow(m.msg.Message(13)) }

func (m Message) HasCode() bool            { return m.msg.HasField(1) }
func (m Message) HasConnectRequest() bool  { return m.msg.HasField(2) }
func (m Message) HasConnectResponse() bool { return m.msg.HasField(3) }
func (m Message) HasBatch() bool           { return m.msg.HasField(4) }
func (m Message) HasChannelOpen() bool     { return m.msg.HasField(10) }
func (m Message) HasChannelClose() bool    { return m.msg.HasField(11) }
func (m Message) HasChannelData() bool     { return m.msg.HasField(12) }
func (m Message) HasChannelWindow() bool   { return m.msg.HasField(13) }

func (m Message) IsEmpty() bool                         { return m.msg.Empty() }
func (m Message) Clone() Message                        { return Message{m.msg.Clone()} }
func (m Message) CloneToArena(a alloc.Arena) Message    { return Message{m.msg.CloneToArena(a)} }
func (m Message) CloneToBuffer(b buffer.Buffer) Message { return Message{m.msg.CloneToBuffer(b)} }
func (m Message) Unwrap() spec.Message                  { return m.msg }

// ConnectRequest

type ConnectRequest struct {
	msg spec.Message
}

func NewConnectRequest(b []byte) ConnectRequest {
	msg := spec.NewMessage(b)
	return ConnectRequest{msg}
}

func NewConnectRequestErr(b []byte) (_ ConnectRequest, err error) {
	msg, err := spec.NewMessageErr(b)
	if err != nil {
		return
	}
	return ConnectRequest{msg}, nil
}

func MakeConnectRequest(msg spec.Message) ConnectRequest {
	return ConnectRequest{msg}
}

func ParseConnectRequest(b []byte) (_ ConnectRequest, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return ConnectRequest{msg}, size, nil
}

func (m ConnectRequest) Versions() spec.TypedList[Version] {
	return spec.NewTypedList(m.msg.FieldRaw(1), ParseVersion)
}
func (m ConnectRequest) Compression() spec.TypedList[ConnectCompression] {
	return spec.NewTypedList(m.msg.FieldRaw(2), ParseConnectCompression)
}

func (m ConnectRequest) HasVersions() bool    { return m.msg.HasField(1) }
func (m ConnectRequest) HasCompression() bool { return m.msg.HasField(2) }

func (m ConnectRequest) IsEmpty() bool         { return m.msg.Empty() }
func (m ConnectRequest) Clone() ConnectRequest { return ConnectRequest{m.msg.Clone()} }
func (m ConnectRequest) CloneToArena(a alloc.Arena) ConnectRequest {
	return ConnectRequest{m.msg.CloneToArena(a)}
}
func (m ConnectRequest) CloneToBuffer(b buffer.Buffer) ConnectRequest {
	return ConnectRequest{m.msg.CloneToBuffer(b)}
}
func (m ConnectRequest) Unwrap() spec.Message { return m.msg }

// ConnectResponse

type ConnectResponse struct {
	msg spec.Message
}

func NewConnectResponse(b []byte) ConnectResponse {
	msg := spec.NewMessage(b)
	return ConnectResponse{msg}
}

func NewConnectResponseErr(b []byte) (_ ConnectResponse, err error) {
	msg, err := spec.NewMessageErr(b)
	if err != nil {
		return
	}
	return ConnectResponse{msg}, nil
}

func MakeConnectResponse(msg spec.Message) ConnectResponse {
	return ConnectResponse{msg}
}

func ParseConnectResponse(b []byte) (_ ConnectResponse, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return ConnectResponse{msg}, size, nil
}

func (m ConnectResponse) Ok() bool           { return m.msg.Bool(1) }
func (m ConnectResponse) Error() spec.String { return m.msg.String(2) }
func (m ConnectResponse) Version() Version   { return NewVersion(m.msg.FieldRaw(10)) }
func (m ConnectResponse) Compression() ConnectCompression {
	return NewConnectCompression(m.msg.FieldRaw(11))
}

func (m ConnectResponse) HasOk() bool          { return m.msg.HasField(1) }
func (m ConnectResponse) HasError() bool       { return m.msg.HasField(2) }
func (m ConnectResponse) HasVersion() bool     { return m.msg.HasField(10) }
func (m ConnectResponse) HasCompression() bool { return m.msg.HasField(11) }

func (m ConnectResponse) IsEmpty() bool          { return m.msg.Empty() }
func (m ConnectResponse) Clone() ConnectResponse { return ConnectResponse{m.msg.Clone()} }
func (m ConnectResponse) CloneToArena(a alloc.Arena) ConnectResponse {
	return ConnectResponse{m.msg.CloneToArena(a)}
}
func (m ConnectResponse) CloneToBuffer(b buffer.Buffer) ConnectResponse {
	return ConnectResponse{m.msg.CloneToBuffer(b)}
}
func (m ConnectResponse) Unwrap() spec.Message { return m.msg }

// ConnectCompression

type ConnectCompression int32

const (
	ConnectCompression_None ConnectCompression = 0
	ConnectCompression_Lz4  ConnectCompression = 1
)

func NewConnectCompression(b []byte) ConnectCompression {
	v, _, _ := encoding.DecodeInt32(b)
	return ConnectCompression(v)
}

func ParseConnectCompression(b []byte) (result ConnectCompression, size int, err error) {
	v, size, err := encoding.DecodeInt32(b)
	if err != nil || size == 0 {
		return
	}
	result = ConnectCompression(v)
	return
}

func WriteConnectCompression(b buffer.Buffer, v ConnectCompression) (int, error) {
	return encoding.EncodeInt32(b, int32(v))
}

func (e ConnectCompression) String() string {
	switch e {
	case ConnectCompression_None:
		return "none"
	case ConnectCompression_Lz4:
		return "lz4"
	}
	return ""
}

// Batch

type Batch struct {
	msg spec.Message
}

func NewBatch(b []byte) Batch {
	msg := spec.NewMessage(b)
	return Batch{msg}
}

func NewBatchErr(b []byte) (_ Batch, err error) {
	msg, err := spec.NewMessageErr(b)
	if err != nil {
		return
	}
	return Batch{msg}, nil
}

func MakeBatch(msg spec.Message) Batch {
	return Batch{msg}
}

func ParseBatch(b []byte) (_ Batch, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return Batch{msg}, size, nil
}

func (m Batch) List() spec.TypedList[Message] {
	return spec.NewTypedList(m.msg.FieldRaw(1), ParseMessage)
}
func (m Batch) HasList() bool                       { return m.msg.HasField(1) }
func (m Batch) IsEmpty() bool                       { return m.msg.Empty() }
func (m Batch) Clone() Batch                        { return Batch{m.msg.Clone()} }
func (m Batch) CloneToArena(a alloc.Arena) Batch    { return Batch{m.msg.CloneToArena(a)} }
func (m Batch) CloneToBuffer(b buffer.Buffer) Batch { return Batch{m.msg.CloneToBuffer(b)} }
func (m Batch) Unwrap() spec.Message                { return m.msg }

// ChannelOpen

type ChannelOpen struct {
	msg spec.Message
}

func NewChannelOpen(b []byte) ChannelOpen {
	msg := spec.NewMessage(b)
	return ChannelOpen{msg}
}

func NewChannelOpenErr(b []byte) (_ ChannelOpen, err error) {
	msg, err := spec.NewMessageErr(b)
	if err != nil {
		return
	}
	return ChannelOpen{msg}, nil
}

func MakeChannelOpen(msg spec.Message) ChannelOpen {
	return ChannelOpen{msg}
}

func ParseChannelOpen(b []byte) (_ ChannelOpen, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return ChannelOpen{msg}, size, nil
}

func (m ChannelOpen) Id() bin.Bin128   { return m.msg.Bin128(1) }
func (m ChannelOpen) Window() int32    { return m.msg.Int32(2) }
func (m ChannelOpen) Data() spec.Bytes { return m.msg.Bytes(3) }

func (m ChannelOpen) HasId() bool     { return m.msg.HasField(1) }
func (m ChannelOpen) HasWindow() bool { return m.msg.HasField(2) }
func (m ChannelOpen) HasData() bool   { return m.msg.HasField(3) }

func (m ChannelOpen) IsEmpty() bool      { return m.msg.Empty() }
func (m ChannelOpen) Clone() ChannelOpen { return ChannelOpen{m.msg.Clone()} }
func (m ChannelOpen) CloneToArena(a alloc.Arena) ChannelOpen {
	return ChannelOpen{m.msg.CloneToArena(a)}
}
func (m ChannelOpen) CloneToBuffer(b buffer.Buffer) ChannelOpen {
	return ChannelOpen{m.msg.CloneToBuffer(b)}
}
func (m ChannelOpen) Unwrap() spec.Message { return m.msg }

// ChannelClose

type ChannelClose struct {
	msg spec.Message
}

func NewChannelClose(b []byte) ChannelClose {
	msg := spec.NewMessage(b)
	return ChannelClose{msg}
}

func NewChannelCloseErr(b []byte) (_ ChannelClose, err error) {
	msg, err := spec.NewMessageErr(b)
	if err != nil {
		return
	}
	return ChannelClose{msg}, nil
}

func MakeChannelClose(msg spec.Message) ChannelClose {
	return ChannelClose{msg}
}

func ParseChannelClose(b []byte) (_ ChannelClose, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return ChannelClose{msg}, size, nil
}

func (m ChannelClose) Id() bin.Bin128   { return m.msg.Bin128(1) }
func (m ChannelClose) Data() spec.Bytes { return m.msg.Bytes(2) }

func (m ChannelClose) HasId() bool   { return m.msg.HasField(1) }
func (m ChannelClose) HasData() bool { return m.msg.HasField(2) }

func (m ChannelClose) IsEmpty() bool       { return m.msg.Empty() }
func (m ChannelClose) Clone() ChannelClose { return ChannelClose{m.msg.Clone()} }
func (m ChannelClose) CloneToArena(a alloc.Arena) ChannelClose {
	return ChannelClose{m.msg.CloneToArena(a)}
}
func (m ChannelClose) CloneToBuffer(b buffer.Buffer) ChannelClose {
	return ChannelClose{m.msg.CloneToBuffer(b)}
}
func (m ChannelClose) Unwrap() spec.Message { return m.msg }

// ChannelData

type ChannelData struct {
	msg spec.Message
}

func NewChannelData(b []byte) ChannelData {
	msg := spec.NewMessage(b)
	return ChannelData{msg}
}

func NewChannelDataErr(b []byte) (_ ChannelData, err error) {
	msg, err := spec.NewMessageErr(b)
	if err != nil {
		return
	}
	return ChannelData{msg}, nil
}

func MakeChannelData(msg spec.Message) ChannelData {
	return ChannelData{msg}
}

func ParseChannelData(b []byte) (_ ChannelData, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return ChannelData{msg}, size, nil
}

func (m ChannelData) Id() bin.Bin128   { return m.msg.Bin128(1) }
func (m ChannelData) Data() spec.Bytes { return m.msg.Bytes(2) }

func (m ChannelData) HasId() bool   { return m.msg.HasField(1) }
func (m ChannelData) HasData() bool { return m.msg.HasField(2) }

func (m ChannelData) IsEmpty() bool      { return m.msg.Empty() }
func (m ChannelData) Clone() ChannelData { return ChannelData{m.msg.Clone()} }
func (m ChannelData) CloneToArena(a alloc.Arena) ChannelData {
	return ChannelData{m.msg.CloneToArena(a)}
}
func (m ChannelData) CloneToBuffer(b buffer.Buffer) ChannelData {
	return ChannelData{m.msg.CloneToBuffer(b)}
}
func (m ChannelData) Unwrap() spec.Message { return m.msg }

// ChannelWindow

type ChannelWindow struct {
	msg spec.Message
}

func NewChannelWindow(b []byte) ChannelWindow {
	msg := spec.NewMessage(b)
	return ChannelWindow{msg}
}

func NewChannelWindowErr(b []byte) (_ ChannelWindow, err error) {
	msg, err := spec.NewMessageErr(b)
	if err != nil {
		return
	}
	return ChannelWindow{msg}, nil
}

func MakeChannelWindow(msg spec.Message) ChannelWindow {
	return ChannelWindow{msg}
}

func ParseChannelWindow(b []byte) (_ ChannelWindow, size int, err error) {
	msg, size, err := spec.ParseMessage(b)
	if err != nil || size == 0 {
		return
	}
	return ChannelWindow{msg}, size, nil
}

func (m ChannelWindow) Id() bin.Bin128 { return m.msg.Bin128(1) }
func (m ChannelWindow) Delta() int32   { return m.msg.Int32(2) }

func (m ChannelWindow) HasId() bool    { return m.msg.HasField(1) }
func (m ChannelWindow) HasDelta() bool { return m.msg.HasField(2) }

func (m ChannelWindow) IsEmpty() bool        { return m.msg.Empty() }
func (m ChannelWindow) Clone() ChannelWindow { return ChannelWindow{m.msg.Clone()} }
func (m ChannelWindow) CloneToArena(a alloc.Arena) ChannelWindow {
	return ChannelWindow{m.msg.CloneToArena(a)}
}
func (m ChannelWindow) CloneToBuffer(b buffer.Buffer) ChannelWindow {
	return ChannelWindow{m.msg.CloneToBuffer(b)}
}
func (m ChannelWindow) Unwrap() spec.Message { return m.msg }

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

func (w MessageWriter) Code(v Code) { spec.WriteField(w.w.Field(1), v, WriteCode) }
func (w MessageWriter) ConnectRequest() ConnectRequestWriter {
	w1 := w.w.Field(2).Message()
	return NewConnectRequestWriterTo(w1)
}
func (w MessageWriter) CopyConnectRequest(v ConnectRequest) error {
	return w.w.Field(2).Any(v.Unwrap().Raw())
}
func (w MessageWriter) ConnectResponse() ConnectResponseWriter {
	w1 := w.w.Field(3).Message()
	return NewConnectResponseWriterTo(w1)
}
func (w MessageWriter) CopyConnectResponse(v ConnectResponse) error {
	return w.w.Field(3).Any(v.Unwrap().Raw())
}
func (w MessageWriter) Batch() BatchWriter {
	w1 := w.w.Field(4).Message()
	return NewBatchWriterTo(w1)
}
func (w MessageWriter) CopyBatch(v Batch) error {
	return w.w.Field(4).Any(v.Unwrap().Raw())
}
func (w MessageWriter) ChannelOpen() ChannelOpenWriter {
	w1 := w.w.Field(10).Message()
	return NewChannelOpenWriterTo(w1)
}
func (w MessageWriter) CopyChannelOpen(v ChannelOpen) error {
	return w.w.Field(10).Any(v.Unwrap().Raw())
}
func (w MessageWriter) ChannelClose() ChannelCloseWriter {
	w1 := w.w.Field(11).Message()
	return NewChannelCloseWriterTo(w1)
}
func (w MessageWriter) CopyChannelClose(v ChannelClose) error {
	return w.w.Field(11).Any(v.Unwrap().Raw())
}
func (w MessageWriter) ChannelData() ChannelDataWriter {
	w1 := w.w.Field(12).Message()
	return NewChannelDataWriterTo(w1)
}
func (w MessageWriter) CopyChannelData(v ChannelData) error {
	return w.w.Field(12).Any(v.Unwrap().Raw())
}
func (w MessageWriter) ChannelWindow() ChannelWindowWriter {
	w1 := w.w.Field(13).Message()
	return NewChannelWindowWriterTo(w1)
}
func (w MessageWriter) CopyChannelWindow(v ChannelWindow) error {
	return w.w.Field(13).Any(v.Unwrap().Raw())
}

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
	return NewMessage(bytes), nil
}

func (w MessageWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// ConnectRequestWriter

type ConnectRequestWriter struct {
	w spec.MessageWriter
}

func NewConnectRequestWriter() ConnectRequestWriter {
	w := spec.NewMessageWriter()
	return ConnectRequestWriter{w}
}

func NewConnectRequestWriterBuffer(b buffer.Buffer) ConnectRequestWriter {
	w := spec.NewMessageWriterBuffer(b)
	return ConnectRequestWriter{w}
}

func NewConnectRequestWriterTo(w spec.MessageWriter) ConnectRequestWriter {
	return ConnectRequestWriter{w}
}

func (w ConnectRequestWriter) Versions() spec.ValueListWriter[Version] {
	w1 := w.w.Field(1).List()
	return spec.NewValueListWriter(w1, WriteVersion)
}
func (w ConnectRequestWriter) Compression() spec.ValueListWriter[ConnectCompression] {
	w1 := w.w.Field(2).List()
	return spec.NewValueListWriter(w1, WriteConnectCompression)
}

func (w ConnectRequestWriter) Merge(msg ConnectRequest) error {
	return w.w.Merge(msg.Unwrap())
}

func (w ConnectRequestWriter) End() error {
	return w.w.End()
}

func (w ConnectRequestWriter) Build() (_ ConnectRequest, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return NewConnectRequest(bytes), nil
}

func (w ConnectRequestWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// ConnectResponseWriter

type ConnectResponseWriter struct {
	w spec.MessageWriter
}

func NewConnectResponseWriter() ConnectResponseWriter {
	w := spec.NewMessageWriter()
	return ConnectResponseWriter{w}
}

func NewConnectResponseWriterBuffer(b buffer.Buffer) ConnectResponseWriter {
	w := spec.NewMessageWriterBuffer(b)
	return ConnectResponseWriter{w}
}

func NewConnectResponseWriterTo(w spec.MessageWriter) ConnectResponseWriter {
	return ConnectResponseWriter{w}
}

func (w ConnectResponseWriter) Ok(v bool)         { w.w.Field(1).Bool(v) }
func (w ConnectResponseWriter) Error(v string)    { w.w.Field(2).String(v) }
func (w ConnectResponseWriter) Version(v Version) { spec.WriteField(w.w.Field(10), v, WriteVersion) }
func (w ConnectResponseWriter) Compression(v ConnectCompression) {
	spec.WriteField(w.w.Field(11), v, WriteConnectCompression)
}

func (w ConnectResponseWriter) Merge(msg ConnectResponse) error {
	return w.w.Merge(msg.Unwrap())
}

func (w ConnectResponseWriter) End() error {
	return w.w.End()
}

func (w ConnectResponseWriter) Build() (_ ConnectResponse, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return NewConnectResponse(bytes), nil
}

func (w ConnectResponseWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// BatchWriter

type BatchWriter struct {
	w spec.MessageWriter
}

func NewBatchWriter() BatchWriter {
	w := spec.NewMessageWriter()
	return BatchWriter{w}
}

func NewBatchWriterBuffer(b buffer.Buffer) BatchWriter {
	w := spec.NewMessageWriterBuffer(b)
	return BatchWriter{w}
}

func NewBatchWriterTo(w spec.MessageWriter) BatchWriter {
	return BatchWriter{w}
}

func (w BatchWriter) List() spec.MessageListWriter[MessageWriter] {
	w1 := w.w.Field(1).List()
	return spec.NewMessageListWriter(w1, NewMessageWriterTo)
}

func (w BatchWriter) Merge(msg Batch) error {
	return w.w.Merge(msg.Unwrap())
}

func (w BatchWriter) End() error {
	return w.w.End()
}

func (w BatchWriter) Build() (_ Batch, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return NewBatch(bytes), nil
}

func (w BatchWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// ChannelOpenWriter

type ChannelOpenWriter struct {
	w spec.MessageWriter
}

func NewChannelOpenWriter() ChannelOpenWriter {
	w := spec.NewMessageWriter()
	return ChannelOpenWriter{w}
}

func NewChannelOpenWriterBuffer(b buffer.Buffer) ChannelOpenWriter {
	w := spec.NewMessageWriterBuffer(b)
	return ChannelOpenWriter{w}
}

func NewChannelOpenWriterTo(w spec.MessageWriter) ChannelOpenWriter {
	return ChannelOpenWriter{w}
}

func (w ChannelOpenWriter) Id(v bin.Bin128) { w.w.Field(1).Bin128(v) }
func (w ChannelOpenWriter) Window(v int32)  { w.w.Field(2).Int32(v) }
func (w ChannelOpenWriter) Data(v []byte)   { w.w.Field(3).Bytes(v) }

func (w ChannelOpenWriter) Merge(msg ChannelOpen) error {
	return w.w.Merge(msg.Unwrap())
}

func (w ChannelOpenWriter) End() error {
	return w.w.End()
}

func (w ChannelOpenWriter) Build() (_ ChannelOpen, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return NewChannelOpen(bytes), nil
}

func (w ChannelOpenWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// ChannelCloseWriter

type ChannelCloseWriter struct {
	w spec.MessageWriter
}

func NewChannelCloseWriter() ChannelCloseWriter {
	w := spec.NewMessageWriter()
	return ChannelCloseWriter{w}
}

func NewChannelCloseWriterBuffer(b buffer.Buffer) ChannelCloseWriter {
	w := spec.NewMessageWriterBuffer(b)
	return ChannelCloseWriter{w}
}

func NewChannelCloseWriterTo(w spec.MessageWriter) ChannelCloseWriter {
	return ChannelCloseWriter{w}
}

func (w ChannelCloseWriter) Id(v bin.Bin128) { w.w.Field(1).Bin128(v) }
func (w ChannelCloseWriter) Data(v []byte)   { w.w.Field(2).Bytes(v) }

func (w ChannelCloseWriter) Merge(msg ChannelClose) error {
	return w.w.Merge(msg.Unwrap())
}

func (w ChannelCloseWriter) End() error {
	return w.w.End()
}

func (w ChannelCloseWriter) Build() (_ ChannelClose, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return NewChannelClose(bytes), nil
}

func (w ChannelCloseWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// ChannelDataWriter

type ChannelDataWriter struct {
	w spec.MessageWriter
}

func NewChannelDataWriter() ChannelDataWriter {
	w := spec.NewMessageWriter()
	return ChannelDataWriter{w}
}

func NewChannelDataWriterBuffer(b buffer.Buffer) ChannelDataWriter {
	w := spec.NewMessageWriterBuffer(b)
	return ChannelDataWriter{w}
}

func NewChannelDataWriterTo(w spec.MessageWriter) ChannelDataWriter {
	return ChannelDataWriter{w}
}

func (w ChannelDataWriter) Id(v bin.Bin128) { w.w.Field(1).Bin128(v) }
func (w ChannelDataWriter) Data(v []byte)   { w.w.Field(2).Bytes(v) }

func (w ChannelDataWriter) Merge(msg ChannelData) error {
	return w.w.Merge(msg.Unwrap())
}

func (w ChannelDataWriter) End() error {
	return w.w.End()
}

func (w ChannelDataWriter) Build() (_ ChannelData, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return NewChannelData(bytes), nil
}

func (w ChannelDataWriter) Unwrap() spec.MessageWriter {
	return w.w
}

// ChannelWindowWriter

type ChannelWindowWriter struct {
	w spec.MessageWriter
}

func NewChannelWindowWriter() ChannelWindowWriter {
	w := spec.NewMessageWriter()
	return ChannelWindowWriter{w}
}

func NewChannelWindowWriterBuffer(b buffer.Buffer) ChannelWindowWriter {
	w := spec.NewMessageWriterBuffer(b)
	return ChannelWindowWriter{w}
}

func NewChannelWindowWriterTo(w spec.MessageWriter) ChannelWindowWriter {
	return ChannelWindowWriter{w}
}

func (w ChannelWindowWriter) Id(v bin.Bin128) { w.w.Field(1).Bin128(v) }
func (w ChannelWindowWriter) Delta(v int32)   { w.w.Field(2).Int32(v) }

func (w ChannelWindowWriter) Merge(msg ChannelWindow) error {
	return w.w.Merge(msg.Unwrap())
}

func (w ChannelWindowWriter) End() error {
	return w.w.End()
}

func (w ChannelWindowWriter) Build() (_ ChannelWindow, err error) {
	bytes, err := w.w.Build()
	if err != nil {
		return
	}
	return NewChannelWindow(bytes), nil
}

func (w ChannelWindowWriter) Unwrap() spec.MessageWriter {
	return w.w
}
