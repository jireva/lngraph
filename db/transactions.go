package db

import (
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	createTransactionQuery = `CREATE (t:Transaction {
		TxHash: $txHash,
		Amount: $amount,
		NumConfirmations: $numConfirmations,
		BlockHash: $blockHash,
		BlockHeight: $blockHeight,
		TimeStamp: $timeStamp,
		TotalFees: $totalFees,
		Addresses: $addresses
	} )`

	relTransactionChannelQuery = `MATCH (t:Transaction),(c:Channel)
	WHERE c.ChanPoint STARTS WITH $txHash AND t.TxHash = $txHash
	CREATE (t)-[r:FUNDED]->(c)`
)

// TransactionsImporter implements a Neo4j importer for transactions.
type TransactionsImporter struct {
	Driver neo4j.Driver
}

// Import gets multiple transaction resources and imports them into Neo4j and
// creates relationships between them and the channel each of them is part of.
func (ti TransactionsImporter) Import(transactions []*lnrpc.Transaction, counter chan int) error {
	session := ti.Driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()
	for i, tx := range transactions {
		// Show the amount as positive.
		var amount int64
		if tx.Amount < 0 {
			amount = -tx.Amount
		} else {
			amount = tx.Amount
		}
		var addresses string
		for _, a := range tx.GetDestAddresses() {
			if addresses != "" {
				addresses += ","
			}
			addresses += a
		}
		values := map[string]interface{}{
			"txHash":           tx.TxHash,
			"amount":           amount,
			"numConfirmations": tx.NumConfirmations,
			"blockHash":        tx.BlockHash,
			"blockHeight":      tx.BlockHeight,
			"timeStamp":        time.Unix(tx.TimeStamp, 0).Format("2006-01-02 03:04"),
			"totalFees":        tx.TotalFees,
			"addresses":        addresses,
		}

		if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			return tx.Run(createTransactionQuery, values)
		}); err != nil {
			return err
		}

		if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			return tx.Run(relTransactionChannelQuery, values)
		}); err != nil {
			return err
		}

		counter <- i
	}
	close(counter)
	return nil
}
