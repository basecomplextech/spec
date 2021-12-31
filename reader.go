package spec

import (
	"github.com/baseone-run/library/u128"
	"github.com/baseone-run/library/u256"
)

// Reader reads a value from a byte slice.
type Reader []byte

// NewReader returns a new value reader.
func NewReader(b []byte) Reader {
	return Reader(b)
}

func (r Reader) ReadType() (Type, Reader, error) {
	v, n, err := readType(r)
	if err != nil {
		return 0, r, err
	}

	return v, r.read(n), err
}

func (r Reader) ReadNil() (bool, Reader, error) {
	v, n, err := readType(r)
	if err != nil {
		return false, r, err
	}

	return v == TypeNil, r.read(n), err
}

func (r Reader) ReadBool() (bool, Reader, error) {
	v, n, err := readBool(r)
	if err != nil {
		return false, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadByte() (byte, Reader, error) {
	v, n, err := readByte(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadInt8() (int8, Reader, error) {
	v, n, err := readInt8(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadInt16() (int16, Reader, error) {
	v, n, err := readInt16(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadInt32() (int32, Reader, error) {
	v, n, err := readInt32(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadInt64() (int64, Reader, error) {
	v, n, err := readInt64(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadUint8() (uint8, Reader, error) {
	v, n, err := readUint8(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadUint16() (uint16, Reader, error) {
	v, n, err := readUint16(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadUint32() (uint32, Reader, error) {
	v, n, err := readUint32(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadUint64() (uint64, Reader, error) {
	v, n, err := readUint64(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadU128() (u128.U128, Reader, error) {
	v, n, err := readU128(r)
	if err != nil {
		return u128.U128{}, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadU256() (u256.U256, Reader, error) {
	v, n, err := readU256(r)
	if err != nil {
		return u256.U256{}, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadFloat32() (float32, Reader, error) {
	v, n, err := readFloat32(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadFloat64() (float64, Reader, error) {
	v, n, err := readFloat64(r)
	if err != nil {
		return 0, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadBytes() ([]byte, Reader, error) {
	v, n, err := readBytes(r)
	if err != nil {
		return nil, r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadString() (string, Reader, error) {
	v, n, err := readString(r)
	if err != nil {
		return "", r, err
	}
	return v, r.read(n), err
}

func (r Reader) ReadList() (ListReader, error) {
	return NewListReader(r)
}

func (r Reader) ReadMessage() (MessageReader, error) {
	return NewMessageReader(r)
}

// private

// read
func (r Reader) read(n int) Reader {
	return r[:len(r)-n]
}
