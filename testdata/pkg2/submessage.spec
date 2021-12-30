import (
    "sub/pkg3"
)

options (
    go_package="github.com/baseone-run/spec/testgen/golang/pkg2"
)

message SubMessage {
    key     string      1;
    value   pkg3.Value  2;
}
