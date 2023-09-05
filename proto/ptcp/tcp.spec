

enum Code {
    UNDEFINED = 0;
    NEW_STREAM = 1;
    CLOSE_STREAM = 2;
    STREAM_MESSAGE = 3;
}

message Message {
    code    Code            1;
    new     NewStream       2;
    close   CloseStream     3;
    message StreamMessage   4;
}

message NewStream {
    id  bin128  1;
}

message CloseStream {
    id  bin128  1;
}

message StreamMessage {
    id      bin128  1;
    data    bytes   2;
}
