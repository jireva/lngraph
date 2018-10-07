package neo4j

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

const (
	indexLightningNodePubKeyQuery = "CREATE INDEX ON :Node(PubKey)"
	indexLightningNodeAliasQuery  = "CREATE INDEX ON :Node(Alias)"
	indexChannelIDQuery           = "CREATE INDEX ON :Channel(ChannelID)"
	indexCapacityQuery            = "CREATE INDEX ON :Channel(Capacity)"
)

// CreateIndexes creates neo4j indexes for all lighning resources.
//
// Note that the indexes are not immediately available, but will be created in
// the background.
func CreateIndexes(conn bolt.Conn) error {
	_, err := conn.ExecPipeline([]string{
		indexLightningNodePubKeyQuery,
		indexLightningNodeAliasQuery,
		indexChannelIDQuery,
		indexCapacityQuery,
	}, nil, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
