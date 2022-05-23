import (
    "sub/pkg3"
)

options (
    go_package="github.com/epochtimeout/spec/lang/testgen/golang/pkg2"
)

message Submessage {
    key     string      1;
    value   pkg3.Value  2;
}
