

message Request {
    calls   []Call  1;
}

message Response {
    status  Status  1;
    result  []Item  2;
}

message Call {
    method  string  1;
    args    []Item  2;
}

message Item {
    name    string  1;
    value   bytes   2;
}

message Status {
    code    string  1;
    message string  2;
}
