package bitcoin

import (
	"encoding/hex"
	"fmt"
)

// ReverseHex reverses a hex string (for converting between Bitcoin internal format and display format)
func ReverseHex(s string) string {
	data, err := hex.DecodeString(s)
	if err != nil {
		return s
	}

	// Reverse the bytes
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

	return hex.EncodeToString(data)
}

// FormatTxID formats a transaction ID for display
func FormatTxID(txid string) string {
	if len(txid) == 64 {
		return txid // Already in hex format
	}
	return txid
}

// IsNullDataOutput checks if an output is OP_RETURN (null data)
func IsNullDataOutput(spk ScriptPubKey) bool {
	return spk.Type == "nulldata" || spk.Type == "null_data"
}

// IsMultisigOutput checks if an output is a multisig script
func IsMultisigOutput(spk ScriptPubKey) bool {
	return spk.Type == "multisig"
}

// GetOutputAddresses safely extracts addresses from an output
func GetOutputAddresses(vout Vout) []string {
	return ExtractAddresses(vout.ScriptPubKey)
}

// FormatBTCAmount formats a BTC amount for display (removes trailing zeros)
func FormatBTCAmount(value float64) string {
	return fmt.Sprintf("%.8f", value)
}

// ValidateTxHash checks if a transaction hash is valid (64 hex characters)
func ValidateTxHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}

	_, err := hex.DecodeString(hash)
	return err == nil
}
