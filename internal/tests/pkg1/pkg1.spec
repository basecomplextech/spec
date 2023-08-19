import (
    "pkg2"
)

options (
    go_package="github.com/basecomplextech/spec/internal/tests/pkg1"
)

message Message {
    bool    bool    1;
    byte    byte    2;

    int16   int16   10;
    int32   int32   11;
    int64   int64   12;

    uint16  uint16  20;
    uint32  uint32  21;
    uint64  uint64  22;

    float32 float32 30;
    float64 float64 31;

    bin64   bin64   40;
    bin128  bin128  41;
    bin256  bin256  42;

    string      string  50;
    bytes1      bytes   51;
    message1    message 52;

    enum1       Enum            60;
    struct1     Struct          61;
    submessage  Submessage      62;
    submessage1 pkg2.Submessage 63;

    ints            []int64             70;
    strings         []string            71;
    structs         []Struct            73;
    submessages     []Submessage        74;
    submessages1    []pkg2.Submessage   75;

    any any 80;
}

struct Struct {
    key     int32;
    value   int32;
}

message Submessage {
    value   string      1;
    next    Submessage  2;
}
