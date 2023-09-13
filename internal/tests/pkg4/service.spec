import (
    "pkg1"
    "pkg2"
)

options (
    go_package="github.com/basecomplextech/spec/internal/tests/pkg4"
)

service Service {
    // Subservice doc comment
    subservice(id bin128 1) (Subservice);

    // Method doc comment.
    method();

    // Method1 doc comment.
    method1(msg string 1);

    // Method2 doc comment.
    method2(a int64 1, b string 2, c bool 3) (a int64 1, b string 2, c bool 3);

    // Method3 doc comment.
    method3(
        a00 bool    1,
        a01 byte    2,

        a10 int16   10,
        a11 int32   11,
        a12 int64   12,
        
        a20 uint16  20,
        a21 uint32  21,
        a22 uint64  22,

        a30 float32 30,
        a31 float64 31,

        a40 bin64   40,
        a41 bin128  41,
        a42 bin256  42,

        a50 string  50,
        a51 bytes   51,
        a52 message 52,

        a60 pkg1.Enum       60,
        a61 pkg1.Struct     61,
        a62 pkg1.Submessage 62,
        a63 pkg2.Submessage 63,

        a70 []int64             70,
        a71 []string            71,
        a72 []pkg1.Struct       72,
        a73 []pkg1.Submessage   73,
        a74 []pkg2.Submessage   74,

        a80 any 75,
    ) (ok bool 1);

    // Result0 doc comment.
    result0() (
        a00 bool    1,
        a01 byte    2,

        a10 int16   10,
        a11 int32   11,
        a12 int64   12,
    );

    // Result1 doc comment.
    result1() (
        a20 uint16  20,
        a21 uint32  21,
        a22 uint64  22,

        a30 float32 30,
        a31 float64 31,
    );

    // Result2 doc comment.
    result2() (
        a40 bin64   40,
        a41 bin128  41,
        a42 bin256  42,
    );

    // Result3 doc comment.
    result3() (
        a50 string  50,
        a51 bytes   51,
        a52 message 52,
    );

    // Result4 doc comment.
    result4() (
        a60 pkg1.Enum           60,
        a61 pkg1.Struct         61,
        a62 pkg1.Submessage     62,
        a63 pkg2.Submessage     63,
    );

    // Result5 doc comment.
    result5() (
        a70 []int64             70,
        a71 []string            72,
        a72 []pkg1.Struct       73,
        a73 []pkg1.Submessage   74,
    );

    // Result6 doc comment.
    result6() (
        a74 []pkg2.Submessage   74,
        a80 any                 80,
    );
}

subservice Subservice {
    hello(msg string 1) (msg string 1);
}
