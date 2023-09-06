

enum Code {
    UNDEFINED = 0;
    NEW_CHANNEL = 1;
    CLOSE_CHANNEL = 2;
    CHANNEL_MESSAGE = 3;
}

message Message {
    code    Code            1;
    new     NewChannel      2;
    close   CloseChannel    3;
    message ChannelMessage  4;
}

message NewChannel {
    id  bin128  1;
}

message CloseChannel {
    id  bin128  1;
}

message ChannelMessage {
    id      bin128  1;
    data    bytes   2;
}
