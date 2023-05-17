package abci

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/interchainio/forum/model"
)

var _ abci.Application = App{}

type App struct {
	DB *model.DB
}

// Info/Query Connection
// Return application info
func (App) Info(_ context.Context, info *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return &abci.ResponseInfo{}, nil
}

func (App) Query(_ context.Context, query *abci.RequestQuery) (*abci.ResponseQuery, error) {
	return &abci.ResponseQuery{}, nil
}

// Mempool Connection
// Validate a tx for the mempool
func (app App) CheckTx(ctx context.Context, checktx *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return &abci.ResponseCheckTx{}, nil
}

// Consensus Connection
// Initialize blockchain w validators/other info from TendermintCore
func (App) InitChain(_ context.Context, initchain *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	return &abci.ResponseInitChain{}, nil
}

func (App) PrepareProposal(_ context.Context, prepareproposal *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	return &abci.ResponsePrepareProposal{}, nil
}

func (App) ProcessProposal(_ context.Context, processproposal *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	return &abci.ResponseProcessProposal{}, nil
}

// Commit the state and return the application Merkle root hash
func (App) Commit(_ context.Context, commit *abci.RequestCommit) (*abci.ResponseCommit, error) {
	return &abci.ResponseCommit{}, nil
}

// Deliver the decided block with its txs to the Application
func (App) FinalizeBlock(_ context.Context, finalizeblock *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return &abci.ResponseFinalizeBlock{}, nil
}

// State Sync Connection
// List available snapshots
func (App) ListSnapshots(_ context.Context, listsnapshot *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	return &abci.ResponseListSnapshots{}, nil
}

func (App) OfferSnapshot(_ context.Context, offersnapshot *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	return &abci.ResponseOfferSnapshot{}, nil
}

func (App) LoadSnapshotChunk(_ context.Context, loadsnapshotchunk *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	return &abci.ResponseLoadSnapshotChunk{}, nil
}

func (App) ApplySnapshotChunk(_ context.Context, applysnapshotchunk *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	return &abci.ResponseApplySnapshotChunk{}, nil
}

func (App) ExtendVote(_ context.Context, extendvote *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
	return &abci.ResponseExtendVote{}, nil
}

func (App) VerifyVoteExtension(_ context.Context, verifyvoteextension *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
	return &abci.ResponseVerifyVoteExtension{}, nil
}
