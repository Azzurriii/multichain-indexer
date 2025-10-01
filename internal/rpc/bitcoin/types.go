package bitcoin

// Types align with Bitcoin Core getblock (verbosity=2) response and getrawtransaction.

type Block struct {
	Hash              string        `json:"hash"`
	Confirmations     int64         `json:"confirmations"`
	Height            int64         `json:"height"`
	Version           int64         `json:"version"`
	VersionHex        string        `json:"versionHex"`
	MerkleRoot        string        `json:"merkleroot"`
	Time              int64         `json:"time"`
	MedianTime        int64         `json:"mediantime"`
	Nonce             uint64        `json:"nonce"`
	Bits              string        `json:"bits"`
	Difficulty        float64       `json:"difficulty"`
	ChainWork         string        `json:"chainwork"`
	NTx               int           `json:"nTx"`
	PreviousBlockHash string        `json:"previousblockhash"`
	NextBlockHash     string        `json:"nextblockhash,omitempty"`
	StrippedSize      int64         `json:"strippedsize"`
	Size              int64         `json:"size"`
	Weight            int64         `json:"weight"`
	Tx                []Transaction `json:"tx"`
}

type Transaction struct {
	Txid     string `json:"txid"`
	Hash     string `json:"hash"`
	Version  int64  `json:"version"`
	Size     int64  `json:"size"`
	Vsize    int64  `json:"vsize"`
	Weight   int64  `json:"weight"`
	Locktime uint64 `json:"locktime"`
	Vin      []Vin  `json:"vin"`
	Vout     []Vout `json:"vout"`
	Hex      string `json:"hex,omitempty"`
	// Block info (only if transaction is in a block)
	BlockHash     string `json:"blockhash,omitempty"`
	Confirmations int64  `json:"confirmations,omitempty"`
	Time          int64  `json:"time,omitempty"`
	BlockTime     int64  `json:"blocktime,omitempty"`
}

type Vin struct {
	Txid        string     `json:"txid,omitempty"`
	Vout        uint32     `json:"vout,omitempty"`
	ScriptSig   *ScriptSig `json:"scriptSig,omitempty"`
	Sequence    uint64     `json:"sequence"`
	TxInWitness []string   `json:"txinwitness,omitempty"`
	Coinbase    string     `json:"coinbase,omitempty"`
	PrevOut     *PrevOut   `json:"prevout,omitempty"` // Only available with verbosity=2 or special RPC
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// PrevOut contains the previous output info (requires special RPC call)
type PrevOut struct {
	Value        float64      `json:"value"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type Vout struct {
	Value        float64      `json:"value"`
	N            uint32       `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int      `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Address   string   `json:"address,omitempty"`   // Single address (most common)
	Addresses []string `json:"addresses,omitempty"` // Multiple addresses (deprecated but still in use)
}

// TxOut is used for getutxo responses
type TxOut struct {
	BestBlock     string       `json:"bestblock"`
	Confirmations int64        `json:"confirmations"`
	Value         float64      `json:"value"`
	ScriptPubKey  ScriptPubKey `json:"scriptPubKey"`
	Coinbase      bool         `json:"coinbase"`
}

// DecodedRawTransaction is the result of decoderawtransaction RPC
type DecodedRawTransaction struct {
	Txid     string `json:"txid"`
	Hash     string `json:"hash"`
	Size     int64  `json:"size"`
	Vsize    int64  `json:"vsize"`
	Weight   int64  `json:"weight"`
	Version  int64  `json:"version"`
	Locktime uint64 `json:"locktime"`
	Vin      []Vin  `json:"vin"`
	Vout     []Vout `json:"vout"`
}
