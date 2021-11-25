package db

import (
	"fmt"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	createLightningNodeQuery = `CREATE (n:Node {
		Alias: $alias,
		PubKey: $pubKey,
		LastUpdate: $lastUpdate,
		Color: $color,
		Addresses: $addresses
	} )`
)

// NodesImporter implements a Neo4j importer for nodes.
type NodesImporter struct {
	Driver neo4j.Driver
}

// Import gets multiple node resources and imports them into Neo4j.
func (ni NodesImporter) Import(nodes []*lnrpc.LightningNode, counter chan int) error {
	session := ni.Driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()
	for i, lnode := range nodes {
		var addresses string
		for _, a := range lnode.Addresses {
			if addresses != "" {
				addresses += ","
			}
			addresses += fmt.Sprintf("%s:%s", a.Network, a.Addr)
		}
		nodeValues := map[string]interface{}{
			"alias":      lnode.Alias,
			"pubKey":     lnode.PubKey,
			"lastUpdate": time.Unix(int64(lnode.LastUpdate), 0).Format("2006-01-02 03:04"),
			"color":      lnode.Color,
			"addresses":  addresses,
		}
		if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			return tx.Run(createLightningNodeQuery, nodeValues)
		}); err != nil {
			return err
		}
		counter <- i
	}
	close(counter)
	return nil
}
