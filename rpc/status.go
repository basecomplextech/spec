// Copyright 2023 Ivan Korobkov. All rights reserved.

package rpc

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
)

const ErrorCode status.Code = "rpc_error"

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

// internal

func parseStatus(s prpc.Status) status.Status {
	code := parseStatusCode(s.Code())
	msg := parseStatusMessage(s.Message())
	return status.New(code, msg)
}

func parseStatusCode(code spec.String) status.Code {
	switch status.Code(code) {

	// General class
	case status.CodeNone:
		return status.CodeNone
	case status.CodeOK:
		return status.CodeOK
	case status.CodeTest:
		return status.CodeTest

	// Error class
	case status.CodeError:
		return status.CodeError
	case status.CodeCorrupted:
		return status.CodeCorrupted
	case status.CodeExternalError:
		return status.CodeExternalError

	// Invalid class
	case status.CodeNotFound:
		return status.CodeNotFound
	case status.CodeForbidden:
		return status.CodeForbidden
	case status.CodeUnauthorized:
		return status.CodeUnauthorized

	// Unavailable class
	case status.CodeClosed:
		return status.CodeClosed
	case status.CodeCancelled:
		return status.CodeCancelled
	case status.CodeRollback:
		return status.CodeRollback
	case status.CodeTimeout:
		return status.CodeTimeout
	case status.CodeUnavailable:
		return status.CodeUnavailable
	case status.CodeUnsupported:
		return status.CodeUnsupported

	// Iteration/streaming class
	case status.CodeEnd:
		return status.CodeEnd
	case status.CodeWait:
		return status.CodeWait

	// RPC class
	case ErrorCode:
		return ErrorCode
	}

	return status.Code(code.Clone())
}

func parseStatusMessage(msg spec.String) string {
	if msg == "" {
		return ""
	}
	return msg.Clone()
}
