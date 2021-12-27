module pkg2;

import (
    "pkg3"
)

message SubMessage {
    key     string      1;
    value   pkg3.Value  2;
}
