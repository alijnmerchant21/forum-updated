package forum

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/abci/types"
	cryptoencoding "github.com/cometbft/cometbft/crypto/encoding"
)

func isBanTx(tx []byte) bool {
	return strings.Contains(string(tx), "username")
}

func (app *ForumApp) updateValidator(v types.ValidatorUpdate) {
	pubkey, err := cryptoencoding.PubKeyFromProto(v.PubKey)
	if err != nil {
		panic(fmt.Errorf("can't decode public key: %w", err))
	}
	key := []byte("val" + string(pubkey.Bytes()))

	// add or update validator
	value := bytes.NewBuffer(make([]byte, 0))
	if err := types.WriteMessage(&v, value); err != nil {
		panic(err)
	}
	if err = app.state.DB.Set(key, value.Bytes()); err != nil {
		panic(err)
	}
	app.valAddrToPubKeyMap[string(pubkey.Address())] = v.PubKey

}

func (app *ForumApp) getValidators() (validators []types.ValidatorUpdate) {

	err := app.state.DB.GetValidators(validators)
	if err != nil {
		panic(err)
	}
	return
}

const (
	CodeTypeOK              uint32 = 0
	CodeTypeEncodingError   uint32 = 1
	CodeTypeInvalidTxFormat uint32 = 2
	CodeTypeBanned          uint32 = 3
)
