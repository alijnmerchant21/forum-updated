package forum

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/alijnmerchant21/forum-updated/model"

	"github.com/cometbft/cometbft/abci/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/dgraph-io/badger/v3"
)

type ForumApp struct {
	abci.BaseApplication
	User *model.User
	DB   *model.DB
	Msg  *model.Message

	stagedTxs    [][]byte
	stagedBanTxs [][]byte
}

func NewForumApp(dbDir string) (*ForumApp, error) {
	db, err := model.NewDB(dbDir)
	if err != nil {
		fmt.Printf("Error initializing database: %s\n", err)
		return nil, err
	}

	user := &model.User{}
	//db := &model.DB{}

	return &ForumApp{
		User:         user,
		DB:           db,
		stagedTxs:    make([][]byte, 0),
		stagedBanTxs: make([][]byte, 0),
	}, nil
}

// Return application info
func (ForumApp) Info(_ context.Context, info *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return &abci.ResponseInfo{}, nil
}

// Query blockchain
func (app ForumApp) Query(ctx context.Context, query *abci.RequestQuery) (*abci.ResponseQuery, error) {
	resp := abci.ResponseQuery{Key: query.Data}

	// Parse sender from query data
	sender := string(query.Data)

	if sender == "history" {
		messages, err := model.FetchHistory(app.DB)
		if err != nil {
			fmt.Println("Error fetching history")
			return nil, err
		}
		if err != nil {
			fmt.Println("error unmarshalling when fetching:", err)
		}

		resp.Log = messages
		resp.Value = []byte(messages)

		return &resp, nil
	}
	// Retrieve all message sent by the sender
	messages, err := model.GetMessagesBySender(app.DB, sender)
	if err != nil {
		fmt.Println("Error in query1")
		return nil, err
	}

	// Convert the messages to JSON and return as query result
	resultBytes, err := json.Marshal(messages)
	if err != nil {
		fmt.Println("Error in query")
		return nil, err
	}

	resp.Log = string(resultBytes)
	resp.Value = resultBytes

	return &resp, nil
}

func (app ForumApp) CheckTx(ctx context.Context, checktx *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {

	// Parse the tx message
	msg, err := model.ParseMessage(checktx.Tx)
	if err != nil {
		fmt.Printf("failed to parse transaction message checktx: %v\n", err)
		return &abci.ResponseCheckTx{Code: 1}, err
	}
	fmt.Println("Searching for sender ... ", msg.Sender)
	u, err := app.DB.FindUserByName(msg.Sender)

	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		fmt.Println("problem in check tx: ", string(checktx.Tx))
		return &types.ResponseCheckTx{Code: 1}, err
	}

	if u != nil && u.Banned {
		err = fmt.Errorf("user is banned")
		return &types.ResponseCheckTx{Code: 1}, err
	}
	fmt.Println("Check tx success for ", msg.Message, " and ", msg.Sender)
	return &types.ResponseCheckTx{Code: 0}, nil
}

// Consensus Connection
// Initialize blockchain w validators/other info from CometBFT
func (ForumApp) InitChain(_ context.Context, initchain *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	return &abci.ResponseInitChain{}, nil
}

func (app *ForumApp) PrepareProposal(_ context.Context, proposal *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	fmt.Println("entered prepareProp")

	// prepare proposal puts the BanTx first, then adds the other transactions
	// ProcessProposal should verify this
	proposedTxs := make([][]byte, 0)
	finalProposal := make([][]byte, 0)
	bannedUsersString := make(map[string]struct{})
	for _, tx := range proposal.Txs {
		msg, err := model.ParseMessage(tx)
		if err == nil {
			if !model.IsCurseWord(msg.Message) {
				proposedTxs = append(proposedTxs, tx)
			} else {
				banTx := model.BanTx{UserName: msg.Sender}
				bannedUsersString[msg.Message] = struct{}{}
				resultBytes, err := json.Marshal(banTx)
				if err == nil {
					finalProposal = append(finalProposal, resultBytes)
				}
			}
		}
	}
	// Need to loop again through the proposed Txs to make sure there is none left by a user that was banned after the tx was accepted
	for _, tx := range proposedTxs {
		// there should be no error here as these are just transactions we have checked and added
		msg, _ := model.ParseMessage(tx)
		if _, ok := bannedUsersString[msg.Sender]; !ok {
			finalProposal = append(finalProposal, tx)
		}
	}
	return &types.ResponsePrepareProposal{Txs: finalProposal}, nil
}

func (ForumApp) ProcessProposal(_ context.Context, processproposal *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	fmt.Println("entered processProp")
	processedBanTxs := false
	bannedUsers := make(map[string]struct{}, 0)
	for _, tx := range processproposal.Txs {
		var parsedBan model.BanTx
		var parsedTx model.Message
		var err error
		if !processedBanTxs && isBanTx(tx) {
			fmt.Println("FoundBanTx")
			err = json.Unmarshal(tx, &parsedBan)
			if err != nil {
				return &types.ResponseProcessProposal{Status: types.ResponseProcessProposal_REJECT}, err
			}
			fmt.Println("Banned user found", string(tx))
			if processedBanTxs {
				// Banning transactions have to come first, cannot have them once we hit the first non user ban tx
				return &types.ResponseProcessProposal{Status: types.ResponseProcessProposal_REJECT}, err
			}
			bannedUsers[parsedBan.UserName] = struct{}{}
		} else {
			_, err = model.ParseMessage(tx)
			if err != nil {
				fmt.Println(string(tx))
				return &types.ResponseProcessProposal{Status: types.ResponseProcessProposal_REJECT}, err
			}
			processedBanTxs = true
			if _, ok := bannedUsers[parsedTx.Sender]; ok {
				// sending us a tx from a banned user
				return &types.ResponseProcessProposal{Status: types.ResponseProcessProposal_REJECT}, err
			}
		}
	}
	return &types.ResponseProcessProposal{Status: types.ResponseProcessProposal_ACCEPT}, nil
}

// Deliver the decided block with its txs to the Application
func (app *ForumApp) FinalizeBlock(_ context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	fmt.Println("entered finalizeBlock")
	// Iterate over Tx in current block

	app.stagedBanTxs = make([][]byte, 0)
	app.stagedTxs = make([][]byte, 0)
	respTxs := make([]*types.ExecTxResult, len(req.Txs))
	for i, tx := range req.Txs {
		var err error
		//Check if it's a banning transaction
		if isBanTx(tx) {
			banTx := new(model.BanTx)
			err = json.Unmarshal(tx, &banTx)
			if err != nil {
				respTxs[i] = &types.ExecTxResult{Code: 2}
			} else {
				respTxs[i] = &types.ExecTxResult{Code: abci.CodeTypeOK}
				app.stagedBanTxs = append(app.stagedBanTxs, tx)
			}

		} else {
			_, err := model.ParseMessage(tx)
			if err != nil {
				respTxs[i] = &types.ExecTxResult{Code: 2}
			} else {
				app.stagedTxs = append(app.stagedTxs, tx)
				// This adds the user to the DB, but the data is not committed nor persisted until Comit is called
				respTxs[i] = &types.ExecTxResult{Code: abci.CodeTypeOK}
			}
		}
	}
	response := &abci.ResponseFinalizeBlock{TxResults: respTxs}
	return response, nil
}

// Commit the state and return the application Merkle root hash
// Here we actually write the staged transactions into the database.
// For details on why it has to be done here, check the Crash recovery section
// of the ABCI spec
func (app ForumApp) Commit(_ context.Context, commit *abci.RequestCommit) (*abci.ResponseCommit, error) {
	banTx := new(model.BanTx)
	for _, tx := range app.stagedBanTxs {
		err := json.Unmarshal(tx, &banTx)
		if err != nil {
			return nil, err
		}
		u, err := app.DB.FindUserByName(banTx.UserName)
		if err == badger.ErrKeyNotFound {
			newUser := model.User{
				Name:      banTx.UserName,
				PubKey:    ed25519.GenPrivKey().PubKey().Bytes(),
				Moderator: false,
				Banned:    true,
			}
			err := app.DB.CreateUser(&newUser)
			if err != nil {
				return nil, err
			}
			u = &newUser
			continue
		}
		if err == nil {
			u.Banned = true
			err = app.DB.UpdateUser(*u)
			fmt.Println("Error updating user :", err)

		} else {
			return nil, err
		}

	}
	fmt.Println("Commit banned txs succeeded")

	for _, tx := range app.stagedTxs {
		fmt.Println("Committing ", string(tx))
		msg, err := model.ParseMessage(tx)
		if err != nil {
			return nil, err
		} else {
			err = model.AddMessage(app.DB, *msg)
			if err != nil {
				return nil, err
			}
		}
	}

	app.DB.GetDB().Sync()
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

func isBanTx(tx []byte) bool {
	return strings.Contains(string(tx), "username")
}
