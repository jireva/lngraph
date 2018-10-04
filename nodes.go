package main

import bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"

const (
	createLightningNodeQuery = `CREATE (n:Node {
		Alias: {alias},
		PubKey: {pubKey},
		LastUpdate: {lastUpdate},
		Color: {color}
	} )`

	indexLightningNodePubKeyQuery = "CREATE INDEX ON :Node(PubKey)"
	indexLightningNodeAliasQuery  = "CREATE INDEX ON :Node(Alias)"
)

// LightningNode represents a Lightning Network node.
type LightningNode struct {
	LastUpdate uint32 `json:"last_update"`
	PubKey     string `json:"pub_key"`
	Alias      string `json:"alias"`
	Addresses  []struct {
		Network string `json:"network,omitempty"`
		Addr    string `json:"addr,omitempty"`
	} `json:"addresses,omitempty"`
	Color string `json:"color"`
}

// create writes a lightning node resource in neo4j.
func (lnode LightningNode) create(conn bolt.Conn) error {
	values := map[string]interface{}{
		"alias":      lnode.Alias,
		"pubKey":     lnode.PubKey,
		"lastUpdate": lnode.LastUpdate,
		"color":      lnode.Color,
	}
	_, err := conn.ExecNeo(createLightningNodeQuery, values)
	if err != nil {
		return err
	}

	return nil
}

// createLightningNodeIndexes indexes lightning nodes.
func createLightningNodeIndexes(conn bolt.Conn) error {
	_, err := conn.ExecPipeline([]string{
		indexLightningNodePubKeyQuery,
		indexLightningNodeAliasQuery,
	}, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
