package forum

import (
	"strings"

	"github.com/cometbft/cometbft/abci/types"
)

func isBanTx(tx []byte) bool {
	return strings.Contains(string(tx), "username")
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
