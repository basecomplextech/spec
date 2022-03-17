import (
    "pkg2"
)

options (
    go_package="github.com/complexl/spec/testgen/golang/pkg1"
)

message Message {
    field_bool    bool      1;
    field_enum    Enum      2;

    field_int8    int8      10;
    field_int16   int16     11;
    field_int32   int32     12;
    field_int64   int64     13;

    field_uint8   uint8     20;
    field_uint16  uint16    21;
    field_uint32  uint32    22;
    field_uint64  uint64    23;

    field_float32 float32   30;
    field_float64 float64   31;

    field_u128  u128    40;
    field_u256  u256    41;

    field_string  string    50;
    field_bytes   bytes     51;
    field_struct  Struct    52;

    msg         Node            60;
    value       Struct          61;
    imported    pkg2.Submessage 62;

    list_ints       []int64             70;
    list_strings    []string            71;
    list_values     []Struct            73;
    list_messages   []Node              74;
    list_imported   []pkg2.Submessage   75;
}

message Node {
    value   string  1;
    next    Node    2;
}

struct Struct {
    key     int32;
    value   int32;
}
