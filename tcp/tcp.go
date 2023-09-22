package tcp

import (
	"errors"
	"io"
	"net"
	"os"

	"github.com/basecomplextech/baselibrary/status"
)

const (
	ProtocolLine = "SpecTCP/1\n"
)

const (
	codeError status.Code = "tcp_error"
)

var (
	statusClientClosed  = status.Closedf("client closed")
	statusConnClosed    = status.Closedf("connection closed")
	statusChannelClosed = status.Closedf("ch closed")
)

func tcpError(err error) status.Status {
	if err == nil {
		return status.OK
	}

	// IO errors
	switch err {
	case io.EOF:
		return status.End
	case io.ErrUnexpectedEOF:
		return status.WrapError(err).WithCode(codeError)
	}

	// Closed/OS errors
	switch {
	case errors.Is(err, net.ErrClosed):
		return status.WrapError(err).WithCode(status.CodeClosed)
	case errors.Is(err, os.ErrDeadlineExceeded):
		return status.WrapError(err).WithCode(status.CodeTimeout)
	}

	// Net and other errors
	ne, ok := (err).(net.Error)
	switch {
	case !ok:
		return status.WrapError(err).WithCode(codeError)
	case ne.Timeout():
		return status.WrapError(err).WithCode(status.CodeTimeout)
	}

	return status.WrapError(err).WithCode(codeError)
}

func tcpErrorf(format string, args ...any) status.Status {
	return status.Newf(codeError, format, args...)
}
