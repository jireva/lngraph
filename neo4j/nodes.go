package neo4j

import (
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

// CreateNode writes a lightning node resource into neo4j.
func CreateNode(conn bolt.Conn, lnode ln.Node) (bolt.Result, error) {
	return conn.ExecNeo(createLightningNodeQuery, map[string]interface{}{
		"alias":      lnode.Alias,
		"pubKey":     lnode.PubKey,
		"lastUpdate": lnode.LastUpdate,
		"color":      lnode.Color,
	})
}
