package neo4j

import (
	"time"

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

// PeersImporter implements a Neo4j importer for peers.
type PeersImporter struct {
	conn bolt.Conn
}

// NewPeersImporter creates a new PeersImporter.
func NewPeersImporter(conn bolt.Conn) PeersImporter {
	return PeersImporter{
		conn: conn,
	}
}

// Import gets multiple peer resources and imports them into Neo4j and
// creates relationships between them and the user's node.
func (pi PeersImporter) Import(peers []ln.Peer, myPubKey string, counter chan int) error {
	for i, peer := range peers {
		// LND uses milliseconds to represent ping time, Go's time.Duration
		// uses milliseconds instead.
		pingTime := time.Duration(peer.PingTime * 1000)
		if _, err := pi.conn.ExecNeo(relPeerQuery, map[string]interface{}{
			"myPubKey":   myPubKey,
			"peerPubKey": peer.PubKey,
			"bytesSent":  peer.BytesSent,
			"bytesRecv":  peer.BytesRecv,
			"satSent":    peer.SatSent,
			"satRecv":    peer.SatRecv,
			"inbound":    peer.Inbound,
			"pingTime":   pingTime.String(),
		}); err != nil {
			return err
		}
		counter <- i
	}
	close(counter)
	return nil
}
