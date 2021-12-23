module "test";

import (
    "module1"
    "module2"
    "github.com/test/library"
    pkg "github.com/test/package"
)

enum TestEnum {
    UNDEFINED   = 0;
    ONE         = 1;
    TWO         = 2;
    THREE       = 3;
    TestEnum    = 10;
}

message TestMessage {
    bool    bool        1;
    enum    TestEnum    2;

    int8    int8    10;
    int16   int16   11;
    int32   int32   12;
    int64   int64   13;

    uint8   uint8   20;
    uint16  uint16  21;
    uint32  uint32  22;
    uint64  uint64  23;

    float32 float32 30;
    float64 float64 31;

    u128    u128    40;
    u256    u256    41;

    string  string  50;
    bytes   bytes   51;

    list        []int64             60;
    messages    []TestSubMessage    61;
    strings     []string            62;
}

message TestSubMessage {
    int8    int8    1;
    int16   int16   2;
    int32   int32   3;
    int64   int64   4;
}

message TestNode {
    Value   string      1;
    Next    *TestNode   2;
}

struct TestStruct {
    index   int64;
    hash    u256;
}
