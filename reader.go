package spec

import (
	"github.com/baseone-run/library/u128"
	"github.com/baseone-run/library/u256"
)

// Reader reads a value from a byte slice.
type Reader struct {
	b []byte
}

// NewReader returns a new value reader.
func NewReader(b []byte) Reader {
	return Reader{b}
}

func (r Reader) ReadType() (Type, error) {
	v, _, err := readType(r.b)
	return v, err
}

func (r Reader) ReadNil() (bool, error) {
	v, _, err := readType(r.b)
	if err != nil {
		return false, err
	}
	return v == TypeNil, nil
}

func (r Reader) ReadBool() (bool, error) {
	v, _, err := readBool(r.b)
	return v, err
}

func (r Reader) ReadByte() (byte, error) {
	v, _, err := readByte(r.b)
	return v, err
}

func (r Reader) ReadInt8() (int8, error) {
	v, _, err := readInt8(r.b)
	return v, err
}

func (r Reader) ReadInt16() (int16, error) {
	v, _, err := readInt16(r.b)
	return v, err
}

func (r Reader) ReadInt32() (int32, error) {
	v, _, err := readInt32(r.b)
	return v, err
}

func (r Reader) ReadInt64() (int64, error) {
	v, _, err := readInt64(r.b)
	return v, err
}

func (r Reader) ReadUint8() (uint8, error) {
	v, _, err := readUint8(r.b)
	return v, err
}

func (r Reader) ReadUint16() (uint16, error) {
	v, _, err := readUint16(r.b)
	return v, err
}

func (r Reader) ReadUint32() (uint32, error) {
	v, _, err := readUint32(r.b)
	return v, err
}

func (r Reader) ReadUint64() (uint64, error) {
	v, _, err := readUint64(r.b)
	return v, err
}

func (r Reader) ReadU128() (u128.U128, error) {
	v, _, err := readU128(r.b)
	return v, err
}

func (r Reader) ReadU256() (u256.U256, error) {
	v, _, err := readU256(r.b)
	return v, err
}

func (r Reader) ReadFloat32() (float32, error) {
	v, _, err := readFloat32(r.b)
	return v, err
}

func (r Reader) ReadFloat64() (float64, error) {
	v, _, err := readFloat64(r.b)
	return v, err
}

func (r Reader) ReadBytes() ([]byte, error) {
	v, _, err := readBytes(r.b)
	return v, err
}

func (r Reader) ReadString() (string, error) {
	v, _, err := readString(r.b)
	return v, err
}

func (r Reader) ReadList() (ListReader, error) {
	return NewListReader(r.b)
}

func (r Reader) ReadMessage() (MessageReader, error) {
	return NewMessageReader(r.b)
}
