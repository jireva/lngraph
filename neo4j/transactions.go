package neo4j

import (
	"time"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/lightningnetwork/lnd/lnrpc"
)

const (
	createTransactionQuery = `CREATE (t:Transaction {
		TxHash: {txHash},
		Amount: {amount},
		NumConfirmations: {numConfirmations},
		BlockHash: {blockHash},
		BlockHeight: {blockHeight},
		TimeStamp: {timeStamp},
		TotalFees: {totalFees},
		Addresses: {addresses}
	} )`

	relTransactionChannelQuery = `MATCH (t:Transaction),(c:Channel)
	WHERE c.ChanPoint STARTS WITH {txHash} AND t.TxHash = {txHash}
	CREATE (t)-[r:FUNDED]->(c)`
)

// TransactionsImporter implements a Neo4j importer for transactions.
type TransactionsImporter struct {
	conn bolt.Conn
}

// NewTransactionsImporter creates a new TransactionsImporter.
func NewTransactionsImporter(conn bolt.Conn) TransactionsImporter {
	return TransactionsImporter{
		conn: conn,
	}
}

// Import gets multiple transaction resources and imports them into Neo4j and
// creates relationships between them and the channel each of them is part of.
func (ti TransactionsImporter) Import(transactions []*lnrpc.Transaction, counter chan int) error {
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

		if _, err := ti.conn.ExecNeo(createTransactionQuery, values); err != nil {
			return err
		}

		if _, err := ti.conn.ExecNeo(relTransactionChannelQuery, values); err != nil {
			return err
		}

		counter <- i
	}
	close(counter)
	return nil
}
