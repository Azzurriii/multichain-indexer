package bitcoin

import (
	"fmt"

	"github.com/fystack/multichain-indexer/pkg/common/constant"
	"github.com/fystack/multichain-indexer/pkg/common/types"
	"github.com/shopspring/decimal"
)

// ParsedTransfer represents a single value transfer extracted from a Bitcoin transaction
type ParsedTransfer struct {
	From       string
	To         string
	Value      decimal.Decimal
	IsCoinbase bool
	IsChange   bool // Heuristic: might be a change output
}

// ExtractTransfers parses a Bitcoin transaction and extracts meaningful transfers
// This handles the complexity of Bitcoin's UTXO model
func (tx *Transaction) ExtractTransfers(networkID string, blockNum, timestamp uint64, fee decimal.Decimal) []types.Transaction {
	var transfers []types.Transaction

	// Handle coinbase transactions (mining rewards)
	if tx.IsCoinbase() {
		for _, vout := range tx.Vout {
			if vout.Value <= 0 {
				continue
			}
			addresses := ExtractAddresses(vout.ScriptPubKey)
			if len(addresses) == 0 {
				continue
			}

			// Coinbase output - miner reward
			for _, addr := range addresses {
				transfers = append(transfers, types.Transaction{
					TxHash:       tx.Txid,
					NetworkId:    networkID,
					BlockNumber:  blockNum,
					FromAddress:  "coinbase", // Special indicator for mined coins
					ToAddress:    addr,
					AssetAddress: "",
					Amount:       decimal.NewFromFloat(vout.Value).String(),
					Type:         constant.TxnTypeMining,
					TxFee:        decimal.Zero, // No fee for coinbase
					Timestamp:    timestamp,
				})
			}
		}
		return transfers
	}

	// Regular transactions: extract all input addresses and output addresses
	inputAddresses := tx.ExtractInputAddresses()
	outputsByAddress := tx.GroupOutputsByAddress()

	// If we have input addresses (spending addresses), create transfers
	if len(inputAddresses) > 0 {
		// Calculate total input value (if available)
		totalInput := decimal.Zero
		for _, vin := range tx.Vin {
			if vin.PrevOut != nil {
				totalInput = totalInput.Add(decimal.NewFromFloat(vin.PrevOut.Value))
			}
		}

		// For each output, create a transfer record
		feeAssigned := false
		for addr, value := range outputsByAddress {
			// Skip if output goes back to one of the input addresses (likely change)
			// Note: This is a heuristic and may not always be accurate
			isChange := false
			for _, inputAddr := range inputAddresses {
				if addr == inputAddr {
					isChange = true
					break
				}
			}

			// Create transfer record
			// Use the first input address as "from" (simplified model)
			fromAddr := inputAddresses[0]
			if len(inputAddresses) > 1 {
				// Multiple inputs - use a composite indicator
				fromAddr = fmt.Sprintf("%s+%d_more", inputAddresses[0], len(inputAddresses)-1)
			}

			txType := constant.TxnTypeTransfer
			if isChange {
				// Optional: mark change outputs differently
				txType = constant.TxnTypeTransfer // Keep as transfer for now
			}

			tr := types.Transaction{
				TxHash:       tx.Txid,
				NetworkId:    networkID,
				BlockNumber:  blockNum,
				FromAddress:  fromAddr,
				ToAddress:    addr,
				AssetAddress: "",
				Amount:       value.String(),
				Type:         txType,
				TxFee:        decimal.Zero,
				Timestamp:    timestamp,
			}

			// Assign fee to first non-change output
			if !feeAssigned && !isChange && !fee.IsZero() {
				tr.TxFee = fee
				feeAssigned = true
			}

			transfers = append(transfers, tr)
		}
	} else {
		// Fallback: if we can't determine inputs, just record outputs
		// This happens when prevout data is not available
		feeAssigned := false
		for addr, value := range outputsByAddress {
			tr := types.Transaction{
				TxHash:       tx.Txid,
				NetworkId:    networkID,
				BlockNumber:  blockNum,
				FromAddress:  "", // Unknown sender
				ToAddress:    addr,
				AssetAddress: "",
				Amount:       value.String(),
				Type:         constant.TxnTypeTransfer,
				TxFee:        decimal.Zero,
				Timestamp:    timestamp,
			}

			if !feeAssigned && !fee.IsZero() {
				tr.TxFee = fee
				feeAssigned = true
			}

			transfers = append(transfers, tr)
		}
	}

	return transfers
}

// IsCoinbase checks if this is a coinbase (mining) transaction
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) > 0 && tx.Vin[0].Coinbase != ""
}

// ExtractInputAddresses gets all unique addresses from transaction inputs
func (tx *Transaction) ExtractInputAddresses() []string {
	addressSet := make(map[string]bool)
	var addresses []string

	for _, vin := range tx.Vin {
		if vin.Coinbase != "" {
			continue // Skip coinbase inputs
		}

		// If we have prevout data (from verbosity=3 or separate RPC call)
		if vin.PrevOut != nil {
			addrs := ExtractAddresses(vin.PrevOut.ScriptPubKey)
			for _, addr := range addrs {
				if !addressSet[addr] {
					addressSet[addr] = true
					addresses = append(addresses, addr)
				}
			}
		}
	}

	return addresses
}

// GroupOutputsByAddress groups outputs by recipient address and sums values
func (tx *Transaction) GroupOutputsByAddress() map[string]decimal.Decimal {
	outputs := make(map[string]decimal.Decimal)

	for _, vout := range tx.Vout {
		if vout.Value <= 0 {
			continue
		}

		addresses := ExtractAddresses(vout.ScriptPubKey)
		if len(addresses) == 0 {
			// Non-standard or OP_RETURN outputs
			continue
		}

		// For each address in the output
		for _, addr := range addresses {
			value := decimal.NewFromFloat(vout.Value)
			if existing, ok := outputs[addr]; ok {
				outputs[addr] = existing.Add(value)
			} else {
				outputs[addr] = value
			}
		}
	}

	return outputs
}

// CalculateTotalOutput sums all output values
func (tx *Transaction) CalculateTotalOutput() decimal.Decimal {
	total := decimal.Zero
	for _, vout := range tx.Vout {
		if vout.Value > 0 {
			total = total.Add(decimal.NewFromFloat(vout.Value))
		}
	}
	return total
}

// CalculateTotalInput sums all input values (requires prevout data)
func (tx *Transaction) CalculateTotalInput() decimal.Decimal {
	total := decimal.Zero
	for _, vin := range tx.Vin {
		if vin.PrevOut != nil && vin.PrevOut.Value > 0 {
			total = total.Add(decimal.NewFromFloat(vin.PrevOut.Value))
		}
	}
	return total
}

// HasPrevOutData checks if the transaction has previous output data for fee calculation
func (tx *Transaction) HasPrevOutData() bool {
	if tx.IsCoinbase() {
		return false
	}
	for _, vin := range tx.Vin {
		if vin.PrevOut != nil {
			return true
		}
	}
	return false
}
