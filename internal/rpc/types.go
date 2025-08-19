package rpc

import (
	"encoding/json"
	"fmt"
)

// Provider states - used to track the health status of blockchain providers
const (
	StateHealthy     = "healthy"     // Provider is responding normally with good performance
	StateDegraded    = "degraded"    // Provider is responding but with slower than expected performance
	StateUnhealthy   = "unhealthy"   // Provider is experiencing errors or very slow responses
	StateBlacklisted = "blacklisted" // Provider is temporarily excluded from use due to persistent issues
)

// Network types - supported blockchain networks
const (
	NetworkEVM     = "evm"     // Ethereum and EVM-compatible chains
	NetworkSolana  = "solana"  // Solana blockchain
	NetworkTron    = "tron"    // Tron blockchain
	NetworkBitcoin = "bitcoin" // Bitcoin blockchain
	NetworkGeneric = "generic" // Generic/unknown blockchain type
)

// Client types - communication protocols used by blockchain providers
const (
	ClientTypeRPC  = "rpc"  // JSON-RPC protocol
	ClientTypeREST = "rest" // REST API protocol
)

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	ID      any    `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	ID      any             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}
