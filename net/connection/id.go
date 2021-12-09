package connection

import (
	"time"

	"github.com/xsuners/mo/sync/atom"
)

var cid = atom.NewInt64(time.Now().UnixNano())

// Gen .
func GenID() int64 {
	return cid.IncrementAndGet()
}
