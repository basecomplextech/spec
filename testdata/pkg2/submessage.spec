import (
    "sub/pkg3"
)

options (
    go_package="github.com/complexl/spec/testgen/golang/pkg2"
)

message SubMessage {
    key     string      1;
    value   pkg3.Value  2;
}
