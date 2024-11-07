options (
    go_package="github.com/basecomplextech/spec/proto/pmpx2"
)

enum Version {
    UNDEFINED = 0;
    VERSION_1_0 = 10;
}

// Message

enum Code {
    UNDEFINED = 0;

    CONNECT_REQUEST = 1;
    CONNECT_RESPONSE = 2;

    CHANNEL_OPEN = 10;
    CHANNEL_CLOSE = 11;
    CHANNEL_DATA = 12;
    CHANNEL_WINDOW = 13;
}

message Message {
    code    Code    1;

    connect_request     ConnectRequest  2;
    connect_response    ConnectResponse 3;

    channel_open    ChannelOpen     10;
    channel_close   ChannelClose    11;
    channel_data    ChannelData     12;
    channel_window  ChannelWindow   13;
}

// Connect

message ConnectRequest {
    versions    []Version               1; // Proposed versions
    compression []ConnectCompression    2; // Proposed compression algorithms
}

message ConnectResponse {
    ok      bool    1;
    error   string  2;

    version     Version             10; // Negotiated version
    compression ConnectCompression  11; // Negotiated compression algorithm
}

enum ConnectCompression {
    NONE = 0;
    LZ4 = 1;
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

message ChannelData {
    id      bin128  1;
    data    bytes   2;
}

message ChannelWindow {
    id      bin128  1;
    delta   int32   2; // Increment write window by delta
}
