package tcp

import (
	"errors"
	"io"
	"net"

	"github.com/basecomplextech/baselibrary/status"
)

const (
	codeError  status.Code = "tcp_error"
	codeClosed status.Code = "tcp_closed"
)

func tcpError(err error) status.Status {
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
