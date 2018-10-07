package neo4j

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
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
func CreateIndexes(conn bolt.Conn) ([]bolt.Result, error) {
	return conn.ExecPipeline([]string{
		indexLightningNodePubKeyQuery,
		indexLightningNodeAliasQuery,
		indexChannelIDQuery,
		indexChannelCapacityQuery,
		indexChannelPointQuery,
		indexTransactionHashQuery,
	}, nil, nil, nil, nil, nil, nil)
}
