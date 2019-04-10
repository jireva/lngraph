package neo4j

import (
	"time"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/xsb/lngraph/ln"
)

const (
	createLightningNodeQuery = `CREATE (n:Node {
		Alias: {alias},
		PubKey: {pubKey},
		LastUpdate: {lastUpdate},
		Color: {color}
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
func (ni NodesImporter) Import(nodes []ln.Node, counter chan int) error {
	for i, lnode := range nodes {
		if _, err := ni.conn.ExecNeo(createLightningNodeQuery, map[string]interface{}{
			"alias":      lnode.Alias,
			"pubKey":     lnode.PubKey,
			"lastUpdate": time.Unix(lnode.LastUpdate, 0).Format("2006-01-02 03:04"),
			"color":      lnode.Color,
		}); err != nil {
			return err
		}
		counter <- i
	}
	close(counter)
	return nil
}
