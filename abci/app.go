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
	cryptoencoding "github.com/cometbft/cometbft/crypto/encoding"
	cryptoproto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	"github.com/cometbft/cometbft/version"
	"github.com/dgraph-io/badger/v3"
)

const ApplicationVersion = 1

type ForumApp struct {
	abci.BaseApplication
	valAddrToPubKeyMap map[string]cryptoproto.PublicKey
	CurseWords         string
	state              AppState
	onGoingBlock       *badger.Txn
}

func NewForumApp(dbDir string, appConfigPath string) (*ForumApp, error) {

	db, err := model.NewDB(dbDir)
	if err != nil {
		fmt.Printf("Error initializing database: %s\n", err)
		return nil, err
	}
	cfg, err := LoadConfig(appConfigPath)
	if err != nil {
		cfg = new(Config)
		cfg.CurseWords = "bad"
	}
	return &ForumApp{
		state:              loadState(db),
		valAddrToPubKeyMap: make(map[string]cryptoproto.PublicKey),
		CurseWords:         cfg.CurseWords,
	}, nil

}

// Return application info
func (app ForumApp) Info(_ context.Context, info *abci.RequestInfo) (*abci.ResponseInfo, error) {

	//Reading the validators from the DB because CometBFT expects the application to have them in memory
	if len(app.valAddrToPubKeyMap) == 0 && app.state.Height > 0 {
		validators := app.getValidators()
		for _, v := range validators {
			pubkey, err := cryptoencoding.PubKeyFromProto(v.PubKey)
			if err != nil {
				panic(fmt.Errorf("can't decode public key: %w", err))
			}
			app.valAddrToPubKeyMap[string(pubkey.Address())] = v.PubKey
		}
	}
	return &abci.ResponseInfo{
		Version:         version.ABCIVersion,
		AppVersion:      ApplicationVersion,
		LastBlockHeight: app.state.Height,

		LastBlockAppHash: app.state.Hash(),
	}, nil
}

// Query blockchain
func (app ForumApp) Query(ctx context.Context, query *abci.RequestQuery) (*abci.ResponseQuery, error) {
	resp := abci.ResponseQuery{Key: query.Data}

	// Parse sender from query data
	sender := string(query.Data)

	if sender == "history" {
		messages, err := model.FetchHistory(app.state.DB)
		if err != nil {
			return nil, err
		}
		resp.Log = messages
		resp.Value = []byte(messages)

		return &resp, nil
	}
	// Retrieve all message sent by the sender
	messages, err := model.GetMessagesBySender(app.state.DB, sender)
	if err != nil {
		return nil, err
	}

	// Convert the messages to JSON and return as query result
	resultBytes, err := json.Marshal(messages)
	if err != nil {
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
		return &abci.ResponseCheckTx{Code: CodeTypeInvalidTxFormat, Log: "Invalid transaction format"}, nil
	}
	fmt.Println("Searching for sender ... ", msg.Sender)
	u, err := app.state.DB.FindUserByName(msg.Sender)

	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			fmt.Println("problem in check tx: ", string(checktx.Tx))
			return &types.ResponseCheckTx{Code: CodeTypeEncodingError}, nil
		}
		fmt.Println("Not found user :", msg.Sender)
	} else {
		if u != nil && u.Banned {
			return &types.ResponseCheckTx{Code: CodeTypeBanned, Log: "User is banned"}, nil
		}
	}
	fmt.Println("Check tx success for ", msg.Message, " and ", msg.Sender)
	return &types.ResponseCheckTx{Code: CodeTypeOK}, nil
}

// Consensus Connection
// Initialize blockchain w validators/other info from CometBFT
func (app ForumApp) InitChain(_ context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	for _, v := range req.Validators {
		app.updateValidator(v)
	}
	appHash := app.state.Hash()

	// This parameter can also be set in the genesis file
	req.ConsensusParams.Abci.VoteExtensionsEnableHeight = 1
	return &abci.ResponseInitChain{ConsensusParams: req.ConsensusParams, AppHash: appHash}, nil
}

func (app *ForumApp) PrepareProposal(_ context.Context, proposal *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	fmt.Println("entered prepareProp")

	voteExtensionCurseWords := app.getWordsFromVe(proposal.LocalLastCommit.Votes)

	// prepare proposal puts the BanTx first, then adds the other transactions
	// ProcessProposal should verify this
	proposedTxs := make([][]byte, 0)
	finalProposal := make([][]byte, 0)
	bannedUsersString := make(map[string]struct{})
	for _, tx := range proposal.Txs {
		msg, err := model.ParseMessage(tx)
		if err == nil {
			proposedTxs = append(proposedTxs, tx)
		}
		// Adding the curse words from vote extensions too
		if !IsCurseWord(msg.Message, voteExtensionCurseWords) {
			proposedTxs = append(proposedTxs, tx)
		} else {
			banTx := model.BanTx{UserName: msg.Sender}
			bannedUsersString[msg.Sender] = struct{}{}
			resultBytes, err := json.Marshal(banTx)
			if err == nil {
				finalProposal = append(finalProposal, resultBytes)
			} else {
				panic(fmt.Errorf("invalid ban transaction in prepareProposal"))
			}
		}
	}

	// Need to loop again through the proposed Txs to make sure there is none left by a user that was banned after the tx was accepted
	for _, tx := range proposedTxs {
		// there should be no error here as these are just transactions we have checked and added
		msg, err := model.ParseMessage(tx)
		if err != nil {
			panic(err)
		}
		if _, ok := bannedUsersString[msg.Sender]; !ok {
			finalProposal = append(finalProposal, tx)
		}
	}
	return &types.ResponsePrepareProposal{Txs: finalProposal}, nil
}

func (ForumApp) ProcessProposal(_ context.Context, processproposal *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	fmt.Println("entered processProp")
	bannedUsers := make(map[string]struct{}, 0)

	finishedBanTxIdx := len(processproposal.Txs)
	for i, tx := range processproposal.Txs {
		if isBanTx(tx) {
			var parsedBan model.BanTx
			err := json.Unmarshal(tx, &parsedBan)
			if err != nil {
				return &types.ResponseProcessProposal{Status: types.ResponseProcessProposal_REJECT}, nil
			}
			bannedUsers[parsedBan.UserName] = struct{}{}
		} else {
			finishedBanTxIdx = i
			break
		}
	}

	for _, tx := range processproposal.Txs[finishedBanTxIdx:] {
		msg, err := model.ParseMessage(tx)
		if err != nil {
			return &types.ResponseProcessProposal{Status: types.ResponseProcessProposal_REJECT}, nil
		}
		if _, ok := bannedUsers[msg.Sender]; ok {
			// sending us a tx from a banned user
			return &types.ResponseProcessProposal{Status: types.ResponseProcessProposal_REJECT}, nil
		}
	}
	return &types.ResponseProcessProposal{Status: types.ResponseProcessProposal_ACCEPT}, nil
}

// Deliver the decided block with its txs to the Application
func (app *ForumApp) FinalizeBlock(_ context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	fmt.Println("entered finalizeBlock")
	// Iterate over Tx in current block
	app.onGoingBlock = app.state.DB.GetDB().NewTransaction(true)
	respTxs := make([]*types.ExecTxResult, len(req.Txs))
	finishedBanTxIdx := len(req.Txs)
	for i, tx := range req.Txs {
		var err error

		if isBanTx(tx) {
			banTx := new(model.BanTx)
			err = json.Unmarshal(tx, &banTx)
			if err != nil {
				respTxs[i] = &types.ExecTxResult{Code: CodeTypeEncodingError}
			} else {
				err := UpdateOrSetUser(app.state.DB, banTx.UserName, true, app.onGoingBlock)
				if err != nil {
					panic(err)
				}
				respTxs[i] = &types.ExecTxResult{Code: CodeTypeOK}
				app.state.Size++
			}
		} else {
			finishedBanTxIdx = i
			break
		}
	}

	for idx, tx := range req.Txs[finishedBanTxIdx:] {
		msg, err := model.ParseMessage(tx)
		i := idx + finishedBanTxIdx
		if err != nil {
			respTxs[i] = &types.ExecTxResult{Code: CodeTypeEncodingError}
		} else {
			// Check if this sender already existed; if not, add the user too
			err := UpdateOrSetUser(app.state.DB, msg.Sender, false, app.onGoingBlock)
			if err != nil {
				panic(err)
			}
			// Add the message for this sender
			message, err := model.AppendToExistingMsgs(app.state.DB, *msg)
			if err != nil {
				return nil, err
			}
			app.onGoingBlock.Set([]byte(msg.Sender+"msg"), []byte(message))
			chatHistory, err := model.AppendToChat(app.state.DB, *msg)
			if err != nil {
				panic(err)
			}
			// Append messages to chat history
			app.onGoingBlock.Set([]byte("history"), []byte(chatHistory))
			// This adds the user to the DB, but the data is not committed nor persisted until Comit is called
			respTxs[i] = &types.ExecTxResult{Code: abci.CodeTypeOK}
			app.state.Size++
		}
	}
	app.state.Height = req.Height

	response := &abci.ResponseFinalizeBlock{TxResults: respTxs, AppHash: app.state.Hash()}
	return response, nil
}

// Commit the state and return the application Merkle root hash
// Here we actually write the staged transactions into the database.
// For details on why it has to be done here, check the Crash recovery section
// of the ABCI spec
func (app ForumApp) Commit(_ context.Context, commit *abci.RequestCommit) (*abci.ResponseCommit, error) {
	if err := app.onGoingBlock.Commit(); err != nil {
		panic(err)
	}
	saveState(&app.state)
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

func (app ForumApp) ExtendVote(_ context.Context, extendvote *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
	fmt.Println("Entered extend vote")

	return &abci.ResponseExtendVote{VoteExtension: []byte(app.CurseWords)}, nil
}

func (app ForumApp) VerifyVoteExtension(_ context.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
	fmt.Println("Entered verify extension") // Will not be called for extensions generated by this validator
	if _, ok := app.valAddrToPubKeyMap[string(req.ValidatorAddress)]; !ok {
		// we do not have a validator with this address mapped; this should never happen
		panic(fmt.Errorf("unknown validator"))
	}
	curseWords := strings.Split(string(req.VoteExtension), "|")
	tmpCurseWordMap := make(map[string]struct{})
	// Verify that we do not have double words and the validator is not trying to cheat us
	for _, word := range curseWords {
		tmpCurseWordMap[word] = struct{}{}
	}
	if len(tmpCurseWordMap) < len(curseWords) {
		// Extension repeats words
		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
	}
	return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
}

func (app *ForumApp) getWordsFromVe(voteExtensions []abci.ExtendedVoteInfo) string {
	curseWordMap := make(map[string]int)
	for _, vote := range voteExtensions {

		// This code gets the curse words and makes sure that we do not add them more than once
		// Thus ensuring each validator only adds one word once

		curseWords := strings.Split(string(vote.GetVoteExtension()), "|")
		for _, word := range curseWords {
			if count, ok := curseWordMap[word]; !ok {
				curseWordMap[word] = 1
			} else {
				curseWordMap[word] = count + 1
			}
		}

	}
	fmt.Println("Processed vote extensions :", curseWordMap)
	majority := len(app.valAddrToPubKeyMap) / 3 // We define the majority to be at least 1/3 of the validators;

	voteExtensionCurseWords := ""
	for word, count := range curseWordMap {
		if count > majority {
			if voteExtensionCurseWords == "" {
				voteExtensionCurseWords = word
			} else {
				voteExtensionCurseWords = voteExtensionCurseWords + "|" + word
			}
		}
	}
	return voteExtensionCurseWords

}
