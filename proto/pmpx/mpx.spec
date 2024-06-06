// Connect

enum Version {
    UNDEFINED = 0;
    Version_1_0 = 10;
}

enum Compress {
    NONE = 0;
    LZ4 = 1;
}

message ConnectRequest {
    versions    []Version   1; // Proposed versions
    compress    []Compress  2; // Proposed compression algorithms
}

message ConnectResponse {
    ok      bool    1;
    error   string  2;

    version     Version     10; // Negotiated version
    compress    Compress    11; // Negotiated compression algorithm
}

// Messages

enum Code {
    UNDEFINED = 0;
    CHANNEL_OPEN = 1;
    CHANNEL_CLOSE = 2;
    CHANNEL_MESSAGE = 3;
    CHANNEL_WINDOW = 4;
}

message Message {
    code        Code                1;
    open        ChannelOpen         2;
    close       ChannelClose        3;
    message     ChannelMessage      4;
    window      ChannelWindow       5;
}

// Channel

message ChannelOpen {
    id      bin128  1;
    window  int32   2; // Channel read/write window, 0 means unlimited
    data    bytes   3; // Optional data
}

message ChannelClose {
    id      bin128  1;
    data    bytes   2;
}

message ChannelMessage {
    id      bin128  1;
    data    bytes   2;
}

message ChannelWindow {
    id      bin128  1;
    delta   int32   2; // Increment write window by delta
}
