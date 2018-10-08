package neo4j

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/xsb/lngraph/ln"
)

const (
	relPeerQuery = `MATCH (n1:Node),(n2:Node)
	WHERE n1.PubKey = {myPubKey} AND n2.PubKey = {peerPubKey}
	CREATE (n1)-[r:PEER {
		BytesSent: {bytesSent},
		BytesRecv: {bytesRecv},
		SatSent: {satSent},
		SatRecv: {satRecv},
		Inbound: {inbound},
		PingTime: {pingTime}
	} ]->(n2)`
)

// CreatePeerRelationship creates relationship between your node and a peer.
func CreatePeerRelationship(conn bolt.Conn, myPubKey string, p ln.Peer) (bolt.Result, error) {
	values := map[string]interface{}{
		"myPubKey":   myPubKey,
		"peerPubKey": p.PubKey,
		"bytesSent":  p.BytesSent,
		"bytesRecv":  p.BytesRecv,
		"satSent":    p.SatSent,
		"satRecv":    p.SatRecv,
		"inbound":    p.Inbound,
		"pingTime":   p.PingTime,
	}
	return conn.ExecNeo(relPeerQuery, values)
}
