package rpc

import (
	"github.com/basecomplextech/baselibrary/status"
)

const (
	ContentType = "application/spec-rpc"
	ErrorCode   = status.Code("rpc_error")
)

// Error returns an RPC error status with the given message.
func Error(msg string) status.Status {
	return status.New(ErrorCode, msg)
}

// Errorf returns an RPC error status with the given message.
func Errorf(format string, a ...any) status.Status {
	return status.Newf(ErrorCode, format, a...)
}

// WrapError returns an RPC error status with the given error.
func WrapError(err error) status.Status {
	return status.WrapError(err).
		WithCode(ErrorCode)
}

// WrapErrorf returns an RPC error status with the given error.
func WrapErrorf(err error, format string, a ...any) status.Status {
	return status.WrapErrorf(err, format, a...).
		WithCode(ErrorCode)
}
