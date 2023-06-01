package model

import (
	"github.com/cometbft/cometbft/crypto/ed25519"
)

type User struct {
	Name          string
	PubKey        ed25519.PubKey `badgerhold:"index"` // this is just a wrapper around bytes
	Moderator     bool
	Banned        bool
	NumMessages   int64
	Version       uint64
	SchemaVersion int
}

func (u *User) IsBanned() bool {
	return u.Banned
}
