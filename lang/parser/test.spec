import (
    "module1"
    "module2"
    "github.com/test/library"
    pkg "github.com/test/package"
)

options (
    go_package = "test/go/package"
    java_package = "test/java/package"
)

enum TestEnum {
    UNDEFINED   = 0;
    ONE         = 1;
    TWO         = 2;
    THREE       = 3;
    TEN         = 10;
}

message TestMessage {
    field_bool    bool      1;
    field_byte    byte      2;
    field_enum    TestEnum  3;

    field_int32   int32     10;
    field_int64   int64     11;

    field_uint32  uint32    20;
    field_uint64  uint64    21;

    field_float32 float32   30;
    field_float64 float64   31;

    field_u128    u128      40;
    field_u256    u256      41;

    field_string  string    50;
    field_bytes   bytes     51;

    field_struct  TestStruct 52;

    list        []int64             60;
    messages    []TestObjectElement 61;
    strings     []string            62;

    imported    pkg.ImportedMessage 70;
}

message TestObjectElement {
    field_int8    int8    1;
    field_int16   int16   2;
    field_int32   int32   3;
    field_int64   int64   4;
}

message TestNode {
    Value   string      1;
    Next    TestNode    2;
}

struct TestStruct {
    index   int64;
    hash    u256;
}
