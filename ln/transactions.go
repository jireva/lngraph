package ln

// Transactions represents blockchain transactions known by the lightning
// network node.
type Transactions struct {
	Transactions []Transaction `json:"transactions"`
}

// Transaction represents a blockchain transaction.
type Transaction struct {
	TxHash           string   `json:"tx_hash"`
	Amount           int64    `json:"amount,string"`
	NumConfirmations int32    `json:"num_confirmations"`
	BlockHash        string   `json:"block_hash"`
	BlockHeight      int32    `json:"block_height"`
	TimeStamp        int64    `json:"time_stamp,string"`
	TotalFees        int64    `json:"total_fees,string"`
	Addresses        []string `json:"dest_addresses,omitempty"`
}
