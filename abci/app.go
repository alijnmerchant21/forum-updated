package forum

import (
	"context"
	"encoding/json"
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
	db, err := model.New(dbDir)
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
	// Parse the tx message from raw bytes
	txBytes := checktx.Tx
	fmt.Printf("raw bytes is %v", checktx.Tx)

	// Parse the tx message
	//msg, err := model.ParseTransactionMessage(string(txBytes))
	msg, err := model.ParseTransactionMessage((txBytes))
	fmt.Printf("msg is: %v", msg)
	if err != nil {
		fmt.Printf("failed to parse transaction message checktx: %v\n", err)
		return &abci.ResponseCheckTx{}, nil
	}

	// Debug print statement to verify the parsed message
	//fmt.Printf("sender: %s\n:", msg.Sender)
	//fmt.Printf("message: %s\n:", msg.Message)

	return &abci.ResponseCheckTx{}, nil
}

// Consensus Connection
// Initialize blockchain w validators/other info from CometBFT
func (ForumApp) InitChain(_ context.Context, initchain *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	return &abci.ResponseInitChain{}, nil
}

func (app *ForumApp) PrepareProposal(_ context.Context, proposal *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	// Extract the public keys from the transaction data
	var pubKeys []ed25519.PubKey
	for _, txBytes := range proposal.Txs {
		//tx, err := model.UnmarshalMessage(txBytes)
		//tx, err := model.ParseTransactionMessage((string(txBytes)))
		tx, err := model.ParseTransactionMessage((txBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction Prepare Proposal: %v", err)
		}

		// Extract the sender's public key from the message
		var sendMsg model.MsgSendMessage
		if err := json.Unmarshal([]byte(tx.Message), &sendMsg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %v", err)
		}

		pubKeys = append(pubKeys, sendMsg.From)
	}

	// Retrieve the raw transaction data using GetRawTxsByPubKeys
	rawTxs, err := app.DB.GetRawTxsByPubKeys(pubKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw txs: %v", err)
	}

	for _, rawTx := range rawTxs {
		//tx, err := model.UnmarshalMessage(rawTx)
		//tx, err := model.ParseTransactionMessage((string(rawTx)))
		tx, err := model.ParseTransactionMessage((rawTx))
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction rawtx: %v", err)
		}

		var sendMsg model.MsgSendMessage
		if err := json.Unmarshal([]byte(tx.Message), &sendMsg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %v", err)
		}

		// Check if the message contains any curse words
		if model.IsCurseWord(sendMsg.Text) {
			// Ban the user
			err := app.messages.SetBan(app.msgSendmessage.From, true)
			if err != nil {
				return nil, fmt.Errorf("failed to ban user: %v", err)
			}

			return nil, fmt.Errorf("message contains curse words")
		}
	}

	return &abci.ResponsePrepareProposal{}, nil
}

func (ForumApp) ProcessProposal(_ context.Context, processproposal *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
}

// Deliver the decided block with its txs to the Application
func (app *ForumApp) FinalizeBlock(_ context.Context, finalizeblock *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	for _, txBytes := range finalizeblock.Txs {
		//tx, err := model.UnmarshalMessage(txBytes)
		//tx, err := model.ParseTransactionMessage((string(txBytes)))
		tx, err := model.ParseTransactionMessage((txBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction finalize: %v", err)
		}

		if err := app.DB.AddMessage(tx); err != nil {
			return nil, fmt.Errorf("failed to add message to database: %v", err)
		}
	}
	return &abci.ResponseFinalizeBlock{}, nil
}

// Commit the state and return the application Merkle root hash
func (ForumApp) Commit(_ context.Context, commit *abci.RequestCommit) (*abci.ResponseCommit, error) {
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
