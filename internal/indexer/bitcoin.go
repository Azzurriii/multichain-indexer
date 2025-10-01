package indexer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fystack/multichain-indexer/internal/rpc"
	"github.com/fystack/multichain-indexer/internal/rpc/bitcoin"
	"github.com/fystack/multichain-indexer/pkg/common/config"
	"github.com/fystack/multichain-indexer/pkg/common/enum"
	"github.com/fystack/multichain-indexer/pkg/common/logger"
	"github.com/fystack/multichain-indexer/pkg/common/types"
)

type BitcoinIndexer struct {
	chainName      string
	config         config.ChainConfig
	failover       *rpc.Failover[bitcoin.BitcoinAPI]
	enrichPrevOuts bool // Whether to enrich transactions with prevout data for fees
}

func NewBitcoinIndexer(chainName string, cfg config.ChainConfig, f *rpc.Failover[bitcoin.BitcoinAPI]) *BitcoinIndexer {
	return &BitcoinIndexer{
		chainName:      chainName,
		config:         cfg,
		failover:       f,
		enrichPrevOuts: true, // Enable by default for production-ready fee calculation
	}
}

func (b *BitcoinIndexer) GetName() string                  { return strings.ToUpper(b.chainName) }
func (b *BitcoinIndexer) GetNetworkType() enum.NetworkType { return enum.NetworkTypeBtc }
func (b *BitcoinIndexer) GetNetworkInternalCode() string   { return b.config.InternalCode }
func (b *BitcoinIndexer) GetNetworkId() string             { return b.config.NetworkId }

func (b *BitcoinIndexer) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	var latest uint64
	err := b.failover.ExecuteWithRetry(ctx, func(c bitcoin.BitcoinAPI) error {
		n, err := c.GetBlockCount(ctx)
		latest = n
		return err
	})
	return latest, err
}

func (b *BitcoinIndexer) GetBlock(ctx context.Context, number uint64) (*types.Block, error) {
	var blk *bitcoin.Block

	// Try to get block with prevout data (verbosity=3) first
	err := b.failover.ExecuteWithRetry(ctx, func(c bitcoin.BitcoinAPI) error {
		hash, err := c.GetBlockHash(ctx, number)
		if err != nil {
			return err
		}

		// Try verbosity=3 first (includes prevout data for fees)
		bb, err := c.GetBlockWithPrevOut(ctx, hash)
		if err == nil {
			blk = bb
			return nil
		}

		// Fallback to verbosity=2
		logger.Debug("GetBlockWithPrevOut failed, falling back to GetBlockVerbose",
			"block", number, "error", err)
		bb, err = c.GetBlockVerbose(ctx, hash)
		if err == nil {
			blk = bb
		}
		return err
	})

	if err != nil {
		return nil, err
	}

	// Enrich with prevout data if not already present and enrichment is enabled
	if b.enrichPrevOuts && !b.blockHasPrevOutData(blk) {
		logger.Debug("Enriching block with prevout data", "block", number)
		err = b.failover.ExecuteWithRetry(ctx, func(c bitcoin.BitcoinAPI) error {
			return c.EnrichBlockWithPrevOuts(ctx, blk)
		})
		if err != nil {
			logger.Warn("Failed to enrich block with prevout data, fees may be incomplete",
				"block", number, "error", err)
		}
	}

	return b.processBlock(blk), nil
}

func (b *BitcoinIndexer) GetBlocks(ctx context.Context, from, to uint64, _ bool) ([]BlockResult, error) {
	if to < from {
		return nil, fmt.Errorf("invalid range")
	}
	nums := make([]uint64, 0, to-from+1)
	for n := from; n <= to; n++ {
		nums = append(nums, n)
	}
	return b.GetBlocksByNumbers(ctx, nums)
}

func (b *BitcoinIndexer) GetBlocksByNumbers(ctx context.Context, nums []uint64) ([]BlockResult, error) {
	if len(nums) == 0 {
		return nil, nil
	}

	blocks := make([]BlockResult, len(nums))
	workers := len(nums)
	if b.config.Throttle.Concurrency > 0 && workers > b.config.Throttle.Concurrency {
		workers = b.config.Throttle.Concurrency
	}

	type job struct {
		num   uint64
		index int
	}

	jobs := make(chan job, workers*2)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				blk, err := b.GetBlock(ctx, j.num)
				blocks[j.index] = BlockResult{Number: j.num, Block: blk}
				if err != nil {
					blocks[j.index].Error = &Error{
						ErrorType: ErrorTypeUnknown,
						Message:   err.Error(),
					}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for i, n := range nums {
			select {
			case <-ctx.Done():
				return
			case jobs <- job{num: n, index: i}:
			}
		}
	}()

	wg.Wait()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var firstErr error
	for _, r := range blocks {
		if r.Error != nil {
			firstErr = fmt.Errorf("block %d: %s", r.Number, r.Error.Message)
			break
		}
	}
	return blocks, firstErr
}

func (b *BitcoinIndexer) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := b.GetLatestBlockNumber(ctx)
	return err == nil
}

// blockHasPrevOutData checks if any transaction in the block has prevout data
func (b *BitcoinIndexer) blockHasPrevOutData(blk *bitcoin.Block) bool {
	if blk == nil {
		return false
	}
	for _, tx := range blk.Tx {
		if tx.HasPrevOutData() {
			return true
		}
	}
	return false
}

// processBlock converts a Bitcoin block to the internal types.Block format
// Following the Tron indexer pattern for comprehensive transaction parsing
func (b *BitcoinIndexer) processBlock(blk *bitcoin.Block) *types.Block {
	if blk == nil {
		return nil
	}

	block := &types.Block{
		Number:       uint64(blk.Height),
		Hash:         blk.Hash,
		ParentHash:   blk.PreviousBlockHash,
		Timestamp:    uint64(blk.Time),
		Transactions: []types.Transaction{},
	}

	// Process each transaction in the block
	for _, tx := range blk.Tx {
		// Calculate fee for this transaction
		fee := tx.CalculateFee()

		// Extract transfers using the comprehensive parsing logic
		transfers := tx.ExtractTransfers(
			b.GetNetworkId(),
			uint64(blk.Height),
			uint64(blk.Time),
			fee,
		)

		block.Transactions = append(block.Transactions, transfers...)
	}

	logger.Debug("Processed Bitcoin block",
		"block", blk.Height,
		"hash", blk.Hash,
		"transactions", len(blk.Tx),
		"transfers", len(block.Transactions))

	return block
}
