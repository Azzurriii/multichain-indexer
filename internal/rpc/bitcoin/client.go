package bitcoin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fystack/multichain-indexer/internal/rpc"
	"github.com/fystack/multichain-indexer/pkg/common/logger"
	"github.com/fystack/multichain-indexer/pkg/ratelimiter"
)

type Client struct {
	*rpc.BaseClient
}

func NewBitcoinClient(url string, auth *rpc.AuthConfig, timeout time.Duration, rl *ratelimiter.PooledRateLimiter) *Client {
	return &Client{
		BaseClient: rpc.NewBaseClient(url, rpc.NetworkBitcoin, rpc.ClientTypeRPC, auth, timeout, rl),
	}
}

func (c *Client) GetBlockCount(ctx context.Context) (uint64, error) {
	resp, err := c.CallRPC(ctx, "getblockcount", []any{})
	if err != nil {
		return 0, err
	}
	var n uint64
	if err := json.Unmarshal(resp.Result, &n); err != nil {
		return 0, fmt.Errorf("decode getblockcount: %w", err)
	}
	return n, nil
}

func (c *Client) GetBlockHash(ctx context.Context, height uint64) (string, error) {
	resp, err := c.CallRPC(ctx, "getblockhash", []any{height})
	if err != nil {
		return "", err
	}
	var hash string
	if err := json.Unmarshal(resp.Result, &hash); err != nil {
		return "", fmt.Errorf("decode getblockhash: %w", err)
	}
	return hash, nil
}

func (c *Client) GetBlockVerbose(ctx context.Context, hash string) (*Block, error) {
	// verbosity=2 returns fully decoded txs
	resp, err := c.CallRPC(ctx, "getblock", []any{hash, 2})
	if err != nil {
		return nil, err
	}
	var blk Block
	if err := json.Unmarshal(resp.Result, &blk); err != nil {
		return nil, fmt.Errorf("decode getblock: %w", err)
	}
	return &blk, nil
}

// GetRawTransaction gets a transaction by hash with optional verbosity
// verbosity: 0=hex, 1=json, 2=json with prevout
func (c *Client) GetRawTransaction(ctx context.Context, txid string, verbosity int) (*Transaction, error) {
	resp, err := c.CallRPC(ctx, "getrawtransaction", []any{txid, verbosity})
	if err != nil {
		return nil, err
	}

	if verbosity == 0 {
		// Returns hex string
		return nil, fmt.Errorf("hex format not supported, use verbosity >= 1")
	}

	var tx Transaction
	if err := json.Unmarshal(resp.Result, &tx); err != nil {
		return nil, fmt.Errorf("decode getrawtransaction: %w", err)
	}
	return &tx, nil
}

// GetBlockWithPrevOut gets a block with verbosity=3 which includes prevout data
// Note: This requires Bitcoin Core 24.0+
func (c *Client) GetBlockWithPrevOut(ctx context.Context, hash string) (*Block, error) {
	// verbosity=3 includes prevout data for spending transactions
	resp, err := c.CallRPC(ctx, "getblock", []any{hash, 3})
	if err != nil {
		// Fallback to verbosity=2 if not supported
		logger.Debug("verbosity=3 not supported, falling back to verbosity=2", "error", err)
		return c.GetBlockVerbose(ctx, hash)
	}

	var blk Block
	if err := json.Unmarshal(resp.Result, &blk); err != nil {
		return nil, fmt.Errorf("decode getblock: %w", err)
	}
	return &blk, nil
}

// EnrichBlockWithPrevOuts enriches a block's transactions with previous output data
// This is needed for fee calculation when verbosity=3 is not supported
func (c *Client) EnrichBlockWithPrevOuts(ctx context.Context, block *Block) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}

	// Process transactions concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrent requests

	for i := range block.Tx {
		tx := &block.Tx[i]

		// Skip coinbase transactions
		if tx.IsCoinbase() {
			continue
		}

		// Skip if already has prevout data
		if tx.HasPrevOutData() {
			continue
		}

		wg.Add(1)
		go func(transaction *Transaction) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			for j := range transaction.Vin {
				vin := &transaction.Vin[j]
				if vin.Txid == "" || vin.Coinbase != "" {
					continue
				}

				// Fetch the previous transaction
				prevTx, err := c.GetRawTransaction(ctx, vin.Txid, 1)
				if err != nil {
					logger.Debug("failed to fetch previous transaction",
						"txid", vin.Txid,
						"error", err)
					continue
				}

				// Extract the referenced output
				if int(vin.Vout) < len(prevTx.Vout) {
					prevOut := prevTx.Vout[vin.Vout]
					vin.PrevOut = &PrevOut{
						Value:        prevOut.Value,
						ScriptPubKey: prevOut.ScriptPubKey,
					}
				}
			}
		}(tx)
	}

	wg.Wait()
	return nil
}

// DecodeRawTransaction decodes a raw transaction hex
func (c *Client) DecodeRawTransaction(ctx context.Context, hexTx string) (*DecodedRawTransaction, error) {
	resp, err := c.CallRPC(ctx, "decoderawtransaction", []any{hexTx})
	if err != nil {
		return nil, err
	}

	var decoded DecodedRawTransaction
	if err := json.Unmarshal(resp.Result, &decoded); err != nil {
		return nil, fmt.Errorf("decode decoderawtransaction: %w", err)
	}
	return &decoded, nil
}

// GetTxOut gets unspent transaction output info
func (c *Client) GetTxOut(ctx context.Context, txid string, vout uint32, includeMempool bool) (*TxOut, error) {
	resp, err := c.CallRPC(ctx, "gettxout", []any{txid, vout, includeMempool})
	if err != nil {
		return nil, err
	}

	// Result is null if output is spent
	if string(resp.Result) == "null" {
		return nil, nil
	}

	var txOut TxOut
	if err := json.Unmarshal(resp.Result, &txOut); err != nil {
		return nil, fmt.Errorf("decode gettxout: %w", err)
	}
	return &txOut, nil
}

// Implement NetworkClient methods passthrough
func (c *Client) GetNetworkType() string { return rpc.NetworkBitcoin }
