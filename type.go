package protocol

// Type specifies an object type.
type Type byte

const (
	TypeNull  Type = 0x00
	TypeTrue  Type = 0x01
	TypeFalse Type = 0x02

	TypeInt8  Type = 0x10
	TypeInt16 Type = 0x11
	TypeInt32 Type = 0x12
	TypeInt64 Type = 0x13

	TypeByte        = TypeUInt8
	TypeUInt8  Type = 0x20
	TypeUInt16 Type = 0x21
	TypeUInt32 Type = 0x22
	TypeUInt64 Type = 0x23

	TypeFloat32 Type = 0x30
	TypeFloat64 Type = 0x31

	TypeBytes  Type = 0x40
	TypeString Type = 0x50
	TypeStruct Type = 0x60
	TypeList   Type = 0x70
)
