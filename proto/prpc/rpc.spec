
// Message

enum MessageType {
    UNDEFINED = 0;
    REQUEST = 1;
    RESPONSE = 2;
    MESSAGE = 3;
    END = 4;
}

message Message {
    type    MessageType 1;
    req     Request     2;
    resp    Response    3;
    msg     bytes       4;
}

// Request

message Request {
    calls   []Call  1;
}

message Call {
    method  string  1;
    input   any     2;
}

// Response

message Response {
    status  Status  1;
    result  any     2;
}

message Status {
    code    string  1;
    message string  2;
}
