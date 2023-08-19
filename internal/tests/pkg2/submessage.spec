import (
    "pkg3/pkg3a"
)

options (
    go_package="github.com/basecomplextech/spec/internal/tests/pkg2"
)

message Submessage {
    key     string      1;
    value   pkg3a.Value 2;
}
