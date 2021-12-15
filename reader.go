package protocol

type reader struct {
	buf readBuffer
}

func read(buf readBuffer) reader {
	return reader{buf: buf}
}

func (r reader) bool() bool {
	t, _ := r.buf.type_()
	if t != TypeTrue {
		return false
	}
	return true
}

func (r reader) byte() byte {
	return r.uint8()
}

func (r reader) int8() int8 {
	t, b := r.buf.type_()
	if t != TypeInt8 {
		return 0
	}

	v, _ := b.int8()
	return v
}

func (r reader) int16() int16 {
	t, b := r.buf.type_()
	if t != TypeInt16 {
		return 0
	}

	v, _ := b.int16()
	return v
}

func (r reader) int32() int32 {
	t, b := r.buf.type_()
	if t != TypeInt32 {
		return 0
	}

	v, _ := b.int32()
	return v
}

func (r reader) int64() int64 {
	t, b := r.buf.type_()
	if t != TypeInt64 {
		return 0
	}

	v, _ := b.int64()
	return v
}

func (r reader) uint8() uint8 {
	t, b := r.buf.type_()
	if t != TypeUInt8 {
		return 0
	}

	v, _ := b.uint8()
	return v
}

func (r reader) uint16() uint16 {
	t, b := r.buf.type_()
	if t != TypeUInt16 {
		return 0
	}

	v, _ := b.uint16()
	return v
}

func (r reader) uint32() uint32 {
	t, b := r.buf.type_()
	if t != TypeUInt32 {
		return 0
	}

	v, _ := b.uint32()
	return v
}

func (r reader) uint64() uint64 {
	t, b := r.buf.type_()
	if t != TypeUInt64 {
		return 0
	}

	v, _ := b.uint64()
	return v
}

func (r reader) float32() float32 {
	t, b := r.buf.type_()
	if t != TypeFloat32 {
		return 0
	}

	v, _ := b.float32()
	return v
}

func (r reader) float64() float64 {
	t, b := r.buf.type_()
	if t != TypeFloat64 {
		return 0
	}

	v, _ := b.float64()
	return v
}

func (r reader) bytes() []byte {
	t, b := r.buf.type_()
	if t != TypeBytes {
		return nil
	}

	size, b := b.size()
	v, _ := b.bytes(size)
	return v
}

func (r reader) string() string {
	t, b := r.buf.type_()
	if t != TypeString {
		return ""
	}

	size, b := b.size()
	v, _ := b.string(size)
	return v
}

func (r reader) list() listReader {
	return readList(r.buf)
}

func (r reader) message() messageReader {
	return readMessage(r.buf)
}

// list

type listReader struct {
	bytes readBuffer

	type_ Type
	size  uint32
	count uint32
	table elementTable
	data  readBuffer
}

func readList(buf readBuffer) listReader {
	type_, b := buf.type_()
	if type_ != TypeList {
		return listReader{}
	}

	size, b := b.listSize()
	count, b := b.listCount()
	table, b := b.listTable(count)
	data, b := b.listData(size, count)

	return listReader{
		bytes: buf,

		type_: type_,
		size:  size,
		count: count,
		table: table,
		data:  data,
	}
}

func (r listReader) element(i int) (reader, bool) {
	elem, ok := r.table.lookup(i)
	if !ok {
		return reader{}, false
	}

	buf := r.data.listElement(elem.offset)
	return read(buf), true
}

// message

type messageReader struct {
	bytes readBuffer

	type_ Type
	size  uint32
	count uint32
	table fieldTable
	data  readBuffer
}

func readMessage(buf readBuffer) messageReader {
	type_, b := buf.type_()
	if type_ != TypeMessage {
		return messageReader{}
	}

	size, b := b.messageSize()
	count, b := b.messageCount()
	table, b := b.messageTable(count)
	data, _ := b.messageData(size, count)

	return messageReader{
		bytes: buf,

		type_: type_,
		size:  size,
		count: count,
		table: table,
		data:  data,
	}
}

func (r messageReader) field(tag uint16) (reader, bool) {
	field, ok := r.table.lookup(tag)
	if !ok {
		return reader{}, false
	}

	buf := r.data.messageField(field.offset)
	return read(buf), true
}
