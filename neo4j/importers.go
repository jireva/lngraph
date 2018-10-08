package neo4j

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/xsb/lngraph/ln"
)

const (
	deleteAllQuery = "MATCH (n) DETACH DELETE n"
)

// DeleteAll will remove existing neo4j nodes and relationships.
//
// The purpose is to allow new data to be entered without creating
// duplications, not to literally delete everything. For example, Indexes
// are not removed as they are useful in the future.
func DeleteAll(conn bolt.Conn) error {
	_, err := conn.ExecNeo(deleteAllQuery, nil)
	if err != nil {
		return err
	}
	return nil
}

// ImportNodes imports lightning nodes into neo4j.
func ImportNodes(conn bolt.Conn, nodes ln.Nodes, counter chan int) error {
	for i, lnode := range nodes {
		if _, err := CreateNode(conn, lnode); err != nil {
			return err
		}
		counter <- i
	}
	close(counter)
	return nil
}

// ImportChannels imports lightning channels into neo4j.
func ImportChannels(conn bolt.Conn, channels ln.Channels, counter chan int) error {
	for i, channel := range channels {
		if _, err := CreateChannel(conn, channel); err != nil {
			return err
		}

		if _, err := CreateChannelNodeRelationships(conn, channel); err != nil {
			return err
		}

		counter <- i
	}
	close(counter)
	return nil
}

// ImportTransactions imports blockchain transactions into neo4j.
func ImportTransactions(conn bolt.Conn, txs []ln.Transaction, counter chan int) error {
	for i, tx := range txs {
		if _, err := CreateTransaction(conn, tx); err != nil {
			return err
		}

		if _, err := CreateTransactionChannelRelationship(conn, tx); err != nil {
			return err
		}

		counter <- i
	}
	close(counter)
	return nil
}

// ImportPeers imports peer relationships into neo4j.
func ImportPeers(conn bolt.Conn, myPubKey string, peers []ln.Peer, counter chan int) error {
	for i, peer := range peers {
		if _, err := CreatePeerRelationship(conn, myPubKey, peer); err != nil {
			return err
		}
		counter <- i
	}
	close(counter)
	return nil
}
