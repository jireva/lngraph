package neo4j

import (
	"fmt"
	"time"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/lightningnetwork/lnd/lnrpc"
)

const (
	createLightningNodeQuery = `CREATE (n:Node {
		Alias: {alias},
		PubKey: {pubKey},
		LastUpdate: {lastUpdate},
		Color: {color},
		Addresses: {addresses}
	} )`
)

// NodesImporter implements a Neo4j importer for nodes.
type NodesImporter struct {
	conn bolt.Conn
}

// NewNodesImporter creates a new NodesImporter.
func NewNodesImporter(conn bolt.Conn) NodesImporter {
	return NodesImporter{
		conn: conn,
	}
}

// Import gets multiple node resources and imports them into Neo4j.
func (ni NodesImporter) Import(nodes []*lnrpc.LightningNode, counter chan int) error {
	for i, lnode := range nodes {
		var addresses string
		for _, a := range lnode.Addresses {
			if addresses != "" {
				addresses += ","
			}
			addresses += fmt.Sprintf("%s:%s", a.Network, a.Addr)
		}
		if _, err := ni.conn.ExecNeo(createLightningNodeQuery, map[string]interface{}{
			"alias":      lnode.Alias,
			"pubKey":     lnode.PubKey,
			"lastUpdate": time.Unix(int64(lnode.LastUpdate), 0).Format("2006-01-02 03:04"),
			"color":      lnode.Color,
			"addresses":  addresses,
		}); err != nil {
			return err
		}
		counter <- i
	}
	close(counter)
	return nil
}
