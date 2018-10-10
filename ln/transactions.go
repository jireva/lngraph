package ln

// Transaction represents a blockchain transaction known by the lightning
// network node.
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

// TransactionsImporter loads and imports chain transaction elements.
type TransactionsImporter interface {
	Import(transactions []Transaction, counter chan int) error
}

// TransactionsHandler handles and imports chain transactions.
type TransactionsHandler struct {
	Transactions []Transaction
	Importer     TransactionsImporter
}

// NewTransactionsHandler creates a new TransactionsHandler.
func NewTransactionsHandler(ti TransactionsImporter) TransactionsHandler {
	return TransactionsHandler{
		Importer: ti,
	}
}

// Load loads transactions into a TransactionsHandler.
func (th *TransactionsHandler) Load(transactions []Transaction) {
	th.Transactions = transactions
}

// Import uses the specific importer method for importing data for persistence.
func (th TransactionsHandler) Import(counter chan int) error {
	return th.Importer.Import(th.Transactions, counter)
}
