package forum

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/alijnmerchant21/forum-updated/model"
	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptoencoding "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/dgraph-io/badger/v3"
)

func isBanTx(tx []byte) bool {
	return strings.Contains(string(tx), "username")
}

func (app *ForumApp) getValidators() (validators []types.ValidatorUpdate) {
	var err error
	validators, err = app.state.DB.GetValidators()
	if err != nil {
		panic(err)
	}
	return
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

func IsCurseWord(word string, curseWords string) bool {
	// Define your list of curse words here
	// For example:
	return strings.Contains(curseWords, word)
}

const (
	CodeTypeOK              uint32 = 0
	CodeTypeEncodingError   uint32 = 1
	CodeTypeInvalidTxFormat uint32 = 2
	CodeTypeBanned          uint32 = 3
)

func UpdateOrSetUser(db *model.DB, uname string, toBan bool, txn *badger.Txn) error {
	var u *model.User
	u, err := db.FindUserByName(uname)
	if errors.Is(err, badger.ErrKeyNotFound) {
		u = new(model.User)
		u.Name = uname
		u.PubKey = ed25519.GenPrivKey().PubKey().Bytes()
		u.Banned = toBan
	} else {
		if err == nil {
			u.Banned = toBan
		} else {
			err = fmt.Errorf("not able to process user")
			return err
		}
	}
	userBytes, err := json.Marshal(u)
	if err != nil {
		fmt.Println("Error marshalling user")
		return err
	}
	return txn.Set([]byte(uname), userBytes)

}
