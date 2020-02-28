package common

type Options struct {
	BindAddr          string `json:"bind_address,omitempty"`
	TLSCertPath       string `json:"tls_cert_path,omitempty"`
	TLSKeyPath        string `json:"tls_cert_key,omitempty"`
	LogLevel          uint64 `json:"log_level,omitempty"`
	LogFile           string `json:"log_file,omitempty"`
	ZcashConfPath     string `json:"zcash_conf,omitempty"`
	NoTLSVeryInsecure bool   `json:"no_tls_very_insecure,omitempty"`
	CacheSize         int    `json:"cache_size,omitempty"`
	RPCUser           string `json:"rpcUser,omitempty"`
	RPCPassword       string `json:"rpcPassword,omitempty"`
	RPCHost           string `json:"rpcHost,omitempty"`
	RPCPort           string `json:"rpcPort,omitempty"`
}

// GetBlockchainInfo return the zcashd rpc `getblockchaininfo` status
// https://zcash-rpc.github.io/getblockchaininfo.html
type GetBlockchainInfo struct {
	Chain                string     `json:"chain"`
	Blocks               int        `json:"blocks"`
	Headers              int        `json:"headers"`
	BestBlockhash        string     `json:"bestblockhash"`
	Difficulty           float64    `json:"difficulty"`
	VerificationProgress float64    `json:"verificationprogress"`
	SizeOnDisk           float64    `json:"size_on_disk"`
	SoftForks            []SoftFork `json:"softforks"`
}

type SoftFork struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type Block struct {
	Hash              string        `json:"hash"`
	Confirmations     int           `json:"confirmations"`
	Size              int           `json:"size"`
	Height            int           `json:"height"`
	Version           int           `json:"version"`
	MerkleRoot        string        `json:"merkleroot"`
	FinalSaplingRoot  string        `json:"finalsaplingroot"`
	TX                []Transaction `json:"tx"`
	Time              int64         `json:"time"`
	Nonce             string        `json:"nonce"`
	Difficulty        float64       `json:"difficulty"`
	PreviousBlockHash string        `json:"previousblockhash"`
	NextBlockHash     string        `json:"nextblockhash"`
}

func (b Block) NumberofTransactions() int {
	return len(b.TX)
}

// TransactionTypes
func (b Block) TransactionTypes() (vin, vout, vjoinsplit int) {
	for _, t := range b.TX {
		vin += len(t.VIn)
		vout += len(t.VOut)
		vjoinsplit += len(t.VJoinSplit)

	}
	return vin, vout, vjoinsplit
}

type Transaction struct {
	Hex          string         `json:"hex"`
	Txid         string         `json:"txid"`
	Version      int            `json:"version"`
	Locktime     int            `json:"locktime"`
	ExpiryHeight int            `json:"expirtheight"`
	VIn          []VInTX        `json:"vin"`
	VOut         []VOutTX       `json:"vout"`
	VJoinSplit   []VJoinSplitTX `json:"vjoinsplit"`
}

type VInTX struct {
	TxID      string `json:"txid"`
	VOut      int    `json:"vout"`
	ScriptSig ScriptSig
	Sequence  int `json:"sequemce"`
}
type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}
type VOutTX struct {
	Value        float64
	N            int
	ScriptPubKey ScriptPubKey
}
type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int      `json:"reqSigs`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}
type VJoinSplitTX struct {
	VPubOldld float64 `json:"vpub_old"`
	VPubNew   float64 `json:"vpub_new"`
}
