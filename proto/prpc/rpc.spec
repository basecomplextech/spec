
// Request

message Request {
    calls   []Call  1;
}

message Call {
    method  string  1;
    args    []Arg   2;
}

message Arg {
    name    string  1;
    value   bytes   2;
}

// Response

message Response {
    status  Status      1;
    results []Result    2;
}

message Status {
    code    string  1;
    message string  2;
}

message Result {
    name    string  1;
    value   bytes   2;
}
