package neo4j

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/xsb/lngraph/ln"
)

const (
	createTransactionQuery = `CREATE (t:Transaction {
		TxHash: {txHash},
		Amount: {amount},
		NumConfirmations: {numConfirmations},
		BlockHash: {blockHash},
		BlockHeight: {blockHeight},
		TimeStamp: {timeStamp},
		TotalFees: {totalFees}
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
func (ti TransactionsImporter) Import(transactions []ln.Transaction, counter chan int) error {
	for i, tx := range transactions {
		values := map[string]interface{}{
			"txHash":           tx.TxHash,
			"amount":           tx.Amount,
			"numConfirmations": tx.NumConfirmations,
			"blockHash":        tx.BlockHash,
			"blockHeight":      tx.BlockHeight,
			"timeStamp":        tx.TimeStamp,
			"totalFees":        tx.TotalFees,
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
