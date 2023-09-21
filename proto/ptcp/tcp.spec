// Connect

enum Version {
    UNDEFINED = 0;
    Version_1_0 = 1;
}

message ConnectRequest {
    versions    []Version   1; // Proposed versions
}

message ConnectResponse {
    ok      bool    1;
    error   string  2;
    version Version 10; // Negotiated version
}

// Messages

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

// Channels

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
