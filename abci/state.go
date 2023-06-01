package forum

import (
	"encoding/binary"

	"github.com/alijnmerchant21/forum-updated/model"
)

type AppState struct {
	DB     *model.DB
	Height int64
	Size   int64
}

var stateKey = "appstate"

func (s AppState) Hash() []byte {
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, s.Size)
	return appHash
}
