package forum

import (
	"context"
	"fmt"

	"github.com/alijnmerchant21/forum-updated/model"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
)

type ForumApp struct {
	abci.BaseApplication
	User *model.User
	DB   *model.DB
	Msg  *model.Message
	db   dbm.DB
}

func NewForumApp(dbDir string) (*ForumApp, error) {
	db, err := model.NewDB(dbDir)
	if err != nil {
		return nil, err
	}

	user := &model.User{}

	return &ForumApp{
		User: user,
		DB:   db,
		db:   dbm.NewMemDB(),
	}, nil
}

// Return application info
func (ForumApp) Info(_ context.Context, info *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return &abci.ResponseInfo{}, nil
}

// Query blockchain
func (app ForumApp) Query(ctx context.Context, query *abci.RequestQuery) (*abci.ResponseQuery, error) {

	return &abci.ResponseQuery{}, nil

}

func (app ForumApp) CheckTx(ctx context.Context, checktx *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {

	// Parse the tx message
	msg, err := model.ParseMessage(checktx.Tx)
	if err != nil {
		fmt.Printf("failed to parse transaction message checktx: %v\n", err)
		return &abci.ResponseCheckTx{}, nil
	}

	u, err := app.DB.FindUserByName(msg.Sender)
	if err != nil {
		fmt.Printf("failed to find user checktx: %v\n", err)
		return &abci.ResponseCheckTx{}, nil
	}

	if u == nil {
		newUser := model.User{
			Name:      msg.Sender,
			PubKey:    ed25519.GenPrivKey().PubKey().Bytes(),
			Moderator: false,
			Banned:    false,
		}

		err := app.DB.CreateUser(&newUser)
		if err != nil {
			fmt.Printf("failed to create user checktx: %v\n", err)
			return &abci.ResponseCheckTx{}, nil
		}

		fmt.Println("User added")
	}

	if u != nil {
		fmt.Println("User exist")
	}

	return &abci.ResponseCheckTx{Code: 0}, nil
}

// Consensus Connection
// Initialize blockchain w validators/other info from CometBFT
func (ForumApp) InitChain(_ context.Context, initchain *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	return &abci.ResponseInitChain{}, nil
}

func (app *ForumApp) PrepareProposal(_ context.Context, proposal *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {

	return &abci.ResponsePrepareProposal{}, nil
}

func (ForumApp) ProcessProposal(_ context.Context, processproposal *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
}

// Deliver the decided block with its txs to the Application
func (app *ForumApp) FinalizeBlock(_ context.Context, finalizeblock *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {

	// Iterate over Tx in current block
	for _, tx := range finalizeblock.Txs {

		// Parse tx
		message, err := model.ParseMessage(tx)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction finalize: %v", err)
		}

		// Add message to DB
		err = model.AddMessage(app.db, *message)
		if err != nil {
			return nil, fmt.Errorf("failed to add message to database: %v", err)
		}

	}
	return &abci.ResponseFinalizeBlock{}, nil
}

// Commit the state and return the application Merkle root hash
func (app ForumApp) Commit(_ context.Context, commit *abci.RequestCommit) (*abci.ResponseCommit, error) {
	if err := app.DB.Commit(); err != nil {
		return nil, err
	}
	return &abci.ResponseCommit{}, nil
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
