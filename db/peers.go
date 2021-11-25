package db

import (
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	relPeerQuery = `MATCH (n1:Node),(n2:Node)
	WHERE n1.PubKey = $myPubKey AND n2.PubKey = $peerPubKey
	CREATE (n1)-[r:PEER {
		BytesSent: $bytesSent,
		BytesRecv: $bytesRecv,
		SatSent: $satSent,
		SatRecv: $satRecv,
		Inbound: $inbound,
		PingTime: $pingTime
	} ]->(n2)`
)

// PeersImporter implements a Neo4j importer for peers.
type PeersImporter struct {
	Driver neo4j.Driver
}

// Import gets multiple peer resources and imports them into Neo4j and
// creates relationships between them and the user's node.
func (pi PeersImporter) Import(peers []*lnrpc.Peer, myPubKey string, counter chan int) error {
	session := pi.Driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()
	for i, peer := range peers {
		// LND uses milliseconds to represent ping time, Go's time.Duration
		// uses seconds instead.
		pingTime := time.Duration(peer.PingTime * 1000)
		peerValues := map[string]interface{}{
			"myPubKey":   myPubKey,
			"peerPubKey": peer.PubKey,
			"bytesSent":  peer.BytesSent,
			"bytesRecv":  peer.BytesRecv,
			"satSent":    peer.SatSent,
			"satRecv":    peer.SatRecv,
			"inbound":    peer.Inbound,
			"pingTime":   pingTime.String(),
		}
		if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			return tx.Run(relPeerQuery, peerValues)
		}); err != nil {
			return err
		}
		counter <- i
	}
	close(counter)
	return nil
}
