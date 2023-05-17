package abci

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"forum/model"

	"github.com/cometbft/cometbft/crypto/ed25519"

	// "github.com/alijnmerchant21/forum-updated/model"
	abci "github.com/cometbft/cometbft/abci/types"
)

//var _ abci.Application = ForumApp{}

type ForumApp struct {
	abci.BaseApplication

	DB             *model.DB
	messages       *model.Messages
	msgSendmessage *model.MsgSendMessage
	user           *model.User
}

func NewForumApp(dbDir string) (*ForumApp, error) {
	db, err := model.New(dbDir)
	if err != nil {
		return nil, err
	}

	messages := model.NewMessages(db)
	user := &model.User{}

	return &ForumApp{
		DB:             db,
		messages:       messages,
		msgSendmessage: &model.MsgSendMessage{},
		user:           user,
	}, nil
}

// Return application info
func (ForumApp) Info(_ context.Context, info *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return &abci.ResponseInfo{}, nil
}

// Query blockchain
func (app ForumApp) Query(ctx context.Context, query *abci.RequestQuery) (*abci.ResponseQuery, error) {
	var response abci.ResponseQuery
	switch query.Path {
	case "/messages/":
		pubkeyBytes, err := base64.StdEncoding.DecodeString(string(query.Data))
		if err != nil {
			return nil, err
		}

		pubkey := ed25519.PubKey(pubkeyBytes)

		messages, err := app.messages.GetMessage(pubkey)
		if err != nil {
			return nil, err
		}

		bytes, err := json.Marshal(messages)
		if err != nil {
			return nil, err
		}

		response.Value = bytes
	default:
		return nil, abci.ErrInvalidLengthTypes
	}

	return &response, nil

}

// Mempool Connection
// Validate a tx for the mempool
func (app ForumApp) CheckTx(ctx context.Context, checktx *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	// Find the user by their public key
	user, err := app.DB.FindUser(app.msgSendmessage.From)
	if err != nil {
		return &abci.ResponseCheckTx{}, err
	}

	// Check if the user is banned
	if app.user.IsBanned() {
		return &abci.ResponseCheckTx{}, fmt.Errorf("user with public key %s is banned", user.PubKey)
	}

	return &abci.ResponseCheckTx{}, nil
}

// Consensus Connection
// Initialize blockchain w validators/other info from TendermintCore
func (ForumApp) InitChain(_ context.Context, initchain *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	return &abci.ResponseInitChain{}, nil
}

func (ForumApp) PrepareProposal(_ context.Context, prepareproposal *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	return &abci.ResponsePrepareProposal{}, nil
}

func (ForumApp) ProcessProposal(_ context.Context, processproposal *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	return &abci.ResponseProcessProposal{}, nil
}

// Commit the state and return the application Merkle root hash
func (ForumApp) Commit(_ context.Context, commit *abci.RequestCommit) (*abci.ResponseCommit, error) {
	return &abci.ResponseCommit{}, nil
}

// Deliver the decided block with its txs to the Application
func (ForumApp) FinalizeBlock(_ context.Context, finalizeblock *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return &abci.ResponseFinalizeBlock{}, nil
}

// State Sync Connection
// List available snapshots
func (ForumApp) ListSnapshots(_ context.Context, listsnapshot *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	return &abci.ResponseListSnapshots{}, nil
}

func (ForumApp) OfferSnapshot(_ context.Context, offersnapshot *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	return &abci.ResponseOfferSnapshot{}, nil
}

func (ForumApp) LoadSnapshotChunk(_ context.Context, loadsnapshotchunk *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	return &abci.ResponseLoadSnapshotChunk{}, nil
}

func (ForumApp) ApplySnapshotChunk(_ context.Context, applysnapshotchunk *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	return &abci.ResponseApplySnapshotChunk{}, nil
}

func (ForumApp) ExtendVote(_ context.Context, extendvote *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
	return &abci.ResponseExtendVote{}, nil
}

func (ForumApp) VerifyVoteExtension(_ context.Context, verifyvoteextension *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
	return &abci.ResponseVerifyVoteExtension{}, nil
}

/*func (app ForumApp) Query(ctx context.Context, query *abci.RequestQuery) (*abci.ResponseQuery, error) {
	// Parse the query string to get the public key
	pubKey := query.Data

	// Find the user by their public key
	user, err := app.DB.FindUser(pubKey)
	if err != nil {
		return &abci.ResponseQuery{}, err
	}

	// Get the message index based on the user's number of messages
	messageIndex := user.NumMessages - 1

	// Get the message at the specified index from the Messages array
	message, err := app.messages.GetMessage(messageIndex)
	if err != nil {
		return &abci.ResponseQuery{}, err
	}

	// Return the message in a ResponseQuery struct
	response := &abci.ResponseQuery{
		Value: []byte(message),
	}
	return response, nil
}*/
