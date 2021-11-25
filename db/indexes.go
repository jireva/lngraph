package db

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	indexLightningNodePubKeyQuery = "CREATE INDEX ON :Node(PubKey)"
	indexLightningNodeAliasQuery  = "CREATE INDEX ON :Node(Alias)"
	indexChannelIDQuery           = "CREATE INDEX ON :Channel(ChannelID)"
	indexChannelCapacityQuery     = "CREATE INDEX ON :Channel(Capacity)"
	indexChannelPointQuery        = "CREATE INDEX ON :Channel(ChanPoint)"
	indexTransactionHashQuery     = "CREATE INDEX ON :Transaction(TxHash)"
)

// CreateIndexes creates neo4j indexes for all lighning resources.
//
// Note that the indexes are not immediately available, but will be created in
// the background.
func CreateIndexes(driver neo4j.Driver) error {
	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(indexLightningNodePubKeyQuery, nil)
	}); err != nil {
		return err
	}
	if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(indexLightningNodeAliasQuery, nil)
	}); err != nil {
		return err
	}
	if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(indexChannelIDQuery, nil)
	}); err != nil {
		return err
	}
	if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(indexChannelCapacityQuery, nil)
	}); err != nil {
		return err
	}
	if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(indexChannelPointQuery, nil)
	}); err != nil {
		return err
	}
	if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(indexTransactionHashQuery, nil)
	}); err != nil {
		return err
	}
	return nil
}
