package rpc

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
)

func ParseStatus(p prpc.Status) status.Status {
	c := p.Code().Unwrap()
	msg := p.Message().Clone()
	var code status.Code

	// Maybe prevent alloc
	switch status.Code(c) {
	default:
		code = status.Code(p.Code().Clone())

	case status.CodeNone:
		code = status.CodeNone
	case status.CodeOK:
		code = status.CodeOK
	case status.CodeTest:
		code = status.CodeTest

	case status.CodeError:
		code = status.CodeError
	case status.CodeFatal:
		code = status.CodeFatal
	case status.CodeCorrupted:
		code = status.CodeCorrupted
	case status.CodeExternalError:
		code = status.CodeExternalError

	case status.CodeNotFound:
		code = status.CodeNotFound
	case status.CodeForbidden:
		code = status.CodeForbidden
	case status.CodeUnauthorized:
		code = status.CodeUnauthorized

	case status.CodeClosed:
		code = status.CodeClosed
	case status.CodeCancelled:
		code = status.CodeCancelled
	case status.CodeRollback:
		code = status.CodeRollback
	case status.CodeTimeout:
		code = status.CodeTimeout
	case status.CodeUnavailable:
		code = status.CodeUnavailable
	case status.CodeUnsupported:
		code = status.CodeUnsupported

	case status.CodeEnd:
		code = status.CodeEnd
	case status.CodeWait:
		code = status.CodeWait
	}

	return status.Status{
		Code:    code,
		Message: msg,
	}
}
