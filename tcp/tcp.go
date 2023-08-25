package tcp

import (
	"errors"
	"io"
	"net"

	"github.com/basecomplextech/baselibrary/status"
)

const (
	readBufferSize      = 1 << 15 // 32kb
	writeBufferSize     = 1 << 15 // 32kb
	streamWriteQueueCap = 1 << 17 // 128kb, max stream write queue capacity
)

const (
	codeError  status.Code = "tcp_error"
	codeClosed status.Code = "tcp_closed"
)

var (
	statusClosed       = status.New(codeClosed, "connection closed")
	statusStreamClosed = status.New(codeClosed, "stream closed")
)

func tcpError(err error) status.Status {
	if err == nil {
		return status.OK
	}

	switch err {
	case io.EOF:
		return status.End
	case io.ErrUnexpectedEOF:
		return status.WrapError(err).WithCode(codeError)
	}

	if errors.Is(err, net.ErrClosed) {
		return status.WrapError(err).WithCode(codeClosed)
	}

	ne, ok := (err).(net.Error)
	switch {
	case !ok:
		return status.WrapError(err).WithCode(codeError)
	case ne.Timeout():
		return status.Timeout
	}
	return status.WrapError(err).WithCode(codeError)
}

func tcpErrorf(format string, args ...any) status.Status {
	return status.Newf(codeError, format, args...)
}
