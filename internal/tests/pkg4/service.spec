import (
    "pkg1"
    "pkg2"
)

options (
    go_package="github.com/basecomplextech/spec/internal/tests/pkg4"
)

service Service {
    // Subservice doc comment
    subservice(id bin128) (sub Subservice);

    // Method doc comment.
    method();

    // Method1 doc comment.
    method1(msg string);

    // Method2 doc comment.
    method2(a int64, b string, c bool) (a int64, b string, c bool);

    // Args doc comment.
    args(
        a00 bool,
        a01 byte,

        a10 int16,
        a11 int32,
        a12 int64,
        
        a20 uint16,
        a21 uint32,
        a22 uint64,

        a30 float32,
        a31 float64,

        a40 bin64,
        a41 bin128,
        a42 bin256,

        a50 string,
        a51 bytes,
        a52 message,

        a60 pkg1.Enum,
        a61 pkg1.Struct,
        a62 pkg1.Submessage,
        a63 pkg2.Submessage,

        a70 []int64,
        a71 []string,
        a72 []pkg1.Struct,
        a73 []pkg1.Submessage,
        a74 []pkg2.Submessage,

        a80 any
    ) (ok bool);

    // Result0 doc comment.
    result0() (
        a00 bool,
        a01 byte,

        a10 int16,
        a11 int32,
        a12 int64
    );

    // Result1 doc comment.
    result1() (
        a20 uint16,
        a21 uint32,
        a22 uint64,

        a30 float32,
        a31 float64
    );

    // Result2 doc comment.
    result2() (
        a40 bin64,
        a41 bin128,
        a42 bin256
    );

    // Result3 doc comment.
    result3() (
        a50 string,
        a51 bytes,
        a52 message
    );

    // Result4 doc comment.
    result4() (
        a60 pkg1.Enum,
        a61 pkg1.Struct,
        a62 pkg1.Submessage,
        a63 pkg2.Submessage
    );

    // Result5 doc comment.
    result5() (
        a70 []int64,
        a71 []string,
        a72 []pkg1.Struct,
        a73 []pkg1.Submessage
    );

    // Result6 doc comment.
    result6() (
        a74 []pkg2.Submessage,
        a80 any
    );
}

subservice Subservice {
    hello(msg string) (msg string);
}
