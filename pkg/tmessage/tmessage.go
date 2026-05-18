package tmessage

import (
	"github.com/gotd/td/tg"
)

type Dialog struct {
	Peer      tg.InputPeerClass
	Messages  []int
	FileNames []string // parallel to Messages: filename from export JSON (may be empty)
}

type ParseSource func() ([]*Dialog, error)

func Parse(src ParseSource) ([]*Dialog, error) {
	return src()
}
