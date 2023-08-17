

message Request {
    calls   []Call  1;
}

message Response {
    status  string  1;
    result  []Arg   2;
}

message Call {
    method  string  1;
    args    []Arg   2;
}

message Arg {
    name    string  1;
    value   bytes   2;
}
