package bitcoin

import (
	"crypto/sha256"
	"errors"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/btcsuite/btcutil/bech32"
)

// AddressType represents the type of Bitcoin address
type AddressType string

const (
	AddressTypeP2PKH          AddressType = "p2pkh"  // Pay to Public Key Hash (Legacy: 1...)
	AddressTypeP2SH           AddressType = "p2sh"   // Pay to Script Hash (3...)
	AddressTypeP2WPKH         AddressType = "p2wpkh" // Pay to Witness Public Key Hash (bc1q...)
	AddressTypeP2WSH          AddressType = "p2wsh"  // Pay to Witness Script Hash (bc1q...)
	AddressTypeP2TR           AddressType = "p2tr"   // Pay to Taproot (bc1p...)
	AddressTypeUnknown        AddressType = "unknown"
	AddressTypeNonStandard    AddressType = "nonstandard"
	AddressTypeNullData       AddressType = "nulldata" // OP_RETURN
	AddressTypeMultisig       AddressType = "multisig"
	AddressTypePubkey         AddressType = "pubkey"
	AddressTypeWitnessV1      AddressType = "witness_v1_taproot"
	AddressTypeWitnessUnknown AddressType = "witness_unknown"
)

// Network prefixes for mainnet
const (
	MainnetP2PKHPrefix byte   = 0x00
	MainnetP2SHPrefix  byte   = 0x05
	MainnetBech32HRP   string = "bc"
)

// Testnet prefixes
const (
	TestnetP2PKHPrefix byte   = 0x6f
	TestnetP2SHPrefix  byte   = 0xc4
	TestnetBech32HRP   string = "tb"
)

// IsValidAddress checks if a Bitcoin address is valid
func IsValidAddress(addr string, isTestnet bool) bool {
	if addr == "" {
		return false
	}

	// Check Bech32 (SegWit) addresses
	if strings.HasPrefix(addr, "bc1") || strings.HasPrefix(addr, "tb1") {
		return IsValidBech32Address(addr, isTestnet)
	}

	// Check Base58 (Legacy/P2SH) addresses
	return IsValidBase58Address(addr, isTestnet)
}

// IsValidBech32Address validates a Bech32-encoded address (SegWit)
func IsValidBech32Address(addr string, isTestnet bool) bool {
	expectedHRP := MainnetBech32HRP
	if isTestnet {
		expectedHRP = TestnetBech32HRP
	}

	hrp, data, err := bech32.Decode(addr)
	if err != nil {
		return false
	}

	if hrp != expectedHRP {
		return false
	}

	if len(data) == 0 {
		return false
	}

	// Convert from 5-bit to 8-bit encoding
	converted, err := bech32.ConvertBits(data[1:], 5, 8, false)
	if err != nil {
		return false
	}

	// Validate witness version and program length
	witnessVersion := data[0]
	witnessProgram := converted

	// Witness v0: 20 bytes for P2WPKH, 32 bytes for P2WSH
	if witnessVersion == 0 {
		if len(witnessProgram) != 20 && len(witnessProgram) != 32 {
			return false
		}
	}

	// Witness v1 (Taproot): must be 32 bytes
	if witnessVersion == 1 {
		if len(witnessProgram) != 32 {
			return false
		}
	}

	return true
}

// IsValidBase58Address validates a Base58-encoded address (Legacy/P2SH)
func IsValidBase58Address(addr string, isTestnet bool) bool {
	decoded := base58.Decode(addr)
	if len(decoded) != 25 {
		return false
	}

	// Extract version, payload, and checksum
	version := decoded[0]
	payload := decoded[:21]
	checksum := decoded[21:]

	// Validate version byte
	validVersion := false
	if isTestnet {
		validVersion = version == TestnetP2PKHPrefix || version == TestnetP2SHPrefix
	} else {
		validVersion = version == MainnetP2PKHPrefix || version == MainnetP2SHPrefix
	}
	if !validVersion {
		return false
	}

	// Verify checksum
	hash := sha256.Sum256(payload)
	hash = sha256.Sum256(hash[:])
	expectedChecksum := hash[:4]

	for i := 0; i < 4; i++ {
		if checksum[i] != expectedChecksum[i] {
			return false
		}
	}

	return true
}

// DetectAddressType determines the type of a Bitcoin address
func DetectAddressType(addr string, isTestnet bool) AddressType {
	if addr == "" {
		return AddressTypeUnknown
	}

	// Bech32 addresses (SegWit)
	if strings.HasPrefix(addr, "bc1") || strings.HasPrefix(addr, "tb1") {
		expectedHRP := MainnetBech32HRP
		if isTestnet {
			expectedHRP = TestnetBech32HRP
		}

		hrp, data, err := bech32.Decode(addr)
		if err != nil {
			return AddressTypeUnknown
		}

		if hrp != expectedHRP || len(data) == 0 {
			return AddressTypeUnknown
		}

		witnessVersion := data[0]
		converted, err := bech32.ConvertBits(data[1:], 5, 8, false)
		if err != nil {
			return AddressTypeUnknown
		}

		switch witnessVersion {
		case 0:
			if len(converted) == 20 {
				return AddressTypeP2WPKH
			} else if len(converted) == 32 {
				return AddressTypeP2WSH
			}
		case 1:
			if len(converted) == 32 {
				return AddressTypeP2TR
			}
		default:
			return AddressTypeWitnessUnknown
		}
	}

	// Base58 addresses
	decoded := base58.Decode(addr)
	if len(decoded) == 25 {
		version := decoded[0]
		if isTestnet {
			if version == TestnetP2PKHPrefix {
				return AddressTypeP2PKH
			} else if version == TestnetP2SHPrefix {
				return AddressTypeP2SH
			}
		} else {
			if version == MainnetP2PKHPrefix {
				return AddressTypeP2PKH
			} else if version == MainnetP2SHPrefix {
				return AddressTypeP2SH
			}
		}
	}

	return AddressTypeUnknown
}

// ExtractAddresses extracts all addresses from a ScriptPubKey
func ExtractAddresses(spk ScriptPubKey) []string {
	var addresses []string

	// Single address (modern format)
	if spk.Address != "" {
		addresses = append(addresses, spk.Address)
	}

	// Multiple addresses (legacy multisig)
	if len(spk.Addresses) > 0 {
		addresses = append(addresses, spk.Addresses...)
	}

	return addresses
}

// NormalizeAddress normalizes a Bitcoin address (removes whitespace, validates format)
func NormalizeAddress(addr string) (string, error) {
	cleaned := strings.TrimSpace(addr)
	if cleaned == "" {
		return "", errors.New("empty address")
	}

	// Basic validation
	if !IsValidAddress(cleaned, false) && !IsValidAddress(cleaned, true) {
		return "", errors.New("invalid bitcoin address")
	}

	return cleaned, nil
}

// ScriptTypeToAddressType maps Bitcoin script types to address types
func ScriptTypeToAddressType(scriptType string) AddressType {
	scriptType = strings.ToLower(scriptType)
	switch scriptType {
	case "pubkeyhash":
		return AddressTypeP2PKH
	case "scripthash":
		return AddressTypeP2SH
	case "witness_v0_keyhash":
		return AddressTypeP2WPKH
	case "witness_v0_scripthash":
		return AddressTypeP2WSH
	case "witness_v1_taproot":
		return AddressTypeP2TR
	case "nulldata", "null_data":
		return AddressTypeNullData
	case "multisig":
		return AddressTypeMultisig
	case "pubkey":
		return AddressTypePubkey
	case "nonstandard":
		return AddressTypeNonStandard
	case "witness_unknown":
		return AddressTypeWitnessUnknown
	default:
		return AddressTypeUnknown
	}
}
