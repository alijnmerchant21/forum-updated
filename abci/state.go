package forum

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/alijnmerchant21/forum-updated/model"
	"github.com/dgraph-io/badger/v2"
)

type AppState struct {
	DB     *model.DB
	Size   int64 `json:"size"`
	Height int64 `json:"height"`
}

var stateKey = "appstate"

func (s AppState) Hash() []byte {
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, s.Size)
	return appHash
}

func loadState(db *model.DB) AppState {
	var state AppState
	state.DB = db
	stateBytes, err := db.Get([]byte(stateKey))
	if err != nil && err != badger.ErrKeyNotFound {
		panic(err)
	}
	if len(stateBytes) == 0 {
		return state
	}
	err = json.Unmarshal(stateBytes, &state)
	fmt.Println("ST:", state)
	state.DB = db
	if err != nil {
		panic(err)
	}
	return state
}

func saveState(state *AppState) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	err = state.DB.Set([]byte(stateKey), stateBytes)
	fmt.Println(state)
	if err != nil {
		panic(err)
	}
}
