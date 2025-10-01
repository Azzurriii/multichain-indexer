package bitcoin

import (
	"context"

	"github.com/fystack/multichain-indexer/internal/rpc"
)

// BitcoinAPI defines the subset of Bitcoin Core RPC we need.
type BitcoinAPI interface {
	rpc.NetworkClient

	GetBlockCount(ctx context.Context) (uint64, error)
	GetBlockHash(ctx context.Context, height uint64) (string, error)
	GetBlockVerbose(ctx context.Context, hash string) (*Block, error)
	GetBlockWithPrevOut(ctx context.Context, hash string) (*Block, error)
	GetRawTransaction(ctx context.Context, txid string, verbosity int) (*Transaction, error)
	EnrichBlockWithPrevOuts(ctx context.Context, block *Block) error
	DecodeRawTransaction(ctx context.Context, hexTx string) (*DecodedRawTransaction, error)
	GetTxOut(ctx context.Context, txid string, vout uint32, includeMempool bool) (*TxOut, error)
}
