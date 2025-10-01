package bitcoin

import (
	"github.com/shopspring/decimal"
)

const (
	// BTC_TO_SATOSHI is the conversion factor from BTC to satoshis
	BTC_TO_SATOSHI = 100_000_000
)

// CalculateFee computes the transaction fee from inputs and outputs
// Fee = Total Input Value - Total Output Value
func (tx *Transaction) CalculateFee() decimal.Decimal {
	// Coinbase transactions have no fee (they create new coins)
	if tx.IsCoinbase() {
		return decimal.Zero
	}

	// Need previous output data to calculate fee
	if !tx.HasPrevOutData() {
		return decimal.Zero
	}

	totalInput := tx.CalculateTotalInput()
	totalOutput := tx.CalculateTotalOutput()

	fee := totalInput.Sub(totalOutput)

	// Fee should not be negative
	if fee.IsNegative() {
		return decimal.Zero
	}

	return fee
}

// CalculateFeeRate calculates the fee rate in satoshis per virtual byte (sat/vB)
func (tx *Transaction) CalculateFeeRate() decimal.Decimal {
	fee := tx.CalculateFee()
	if fee.IsZero() || tx.Vsize == 0 {
		return decimal.Zero
	}

	// Convert fee to satoshis
	feeSatoshis := fee.Mul(decimal.NewFromInt(BTC_TO_SATOSHI))

	// Divide by vsize to get sat/vB
	vsize := decimal.NewFromInt(tx.Vsize)
	return feeSatoshis.Div(vsize)
}

// BTCToSatoshi converts BTC to satoshis
func BTCToSatoshi(btc decimal.Decimal) decimal.Decimal {
	return btc.Mul(decimal.NewFromInt(BTC_TO_SATOSHI))
}

// SatoshiToBTC converts satoshis to BTC
func SatoshiToBTC(satoshi decimal.Decimal) decimal.Decimal {
	return satoshi.Div(decimal.NewFromInt(BTC_TO_SATOSHI))
}

// EstimateFeePriority estimates fee priority based on fee rate
// Returns: "high", "medium", "low", or "very_low"
func EstimateFeePriority(feeRateSatPerVByte decimal.Decimal) string {
	// These thresholds are approximate and can vary based on network conditions
	if feeRateSatPerVByte.GreaterThanOrEqual(decimal.NewFromInt(100)) {
		return "high"
	} else if feeRateSatPerVByte.GreaterThanOrEqual(decimal.NewFromInt(50)) {
		return "medium"
	} else if feeRateSatPerVByte.GreaterThanOrEqual(decimal.NewFromInt(10)) {
		return "low"
	}
	return "very_low"
}
