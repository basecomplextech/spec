package mpx

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
)

type internalConn interface {
	write(msg pmpx.Message) status.Status
}
