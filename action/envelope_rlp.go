package action

import "github.com/iotexproject/go-pkgs/hash"

type EnvelopRLP struct {
	Envelope
	chainID uint32
	rawHash hash.Hash256
	hash hash.Hash256
}
