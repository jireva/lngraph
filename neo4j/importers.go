package neo4j

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

const (
	deleteAllQuery = "MATCH (n) DETACH DELETE n"
)

// NewConnection creates a bolt protocol connection to Neo4j.
func NewConnection(url string) (bolt.Conn, error) {
	driver := bolt.NewDriver()
	conn, err := driver.OpenNeo(url)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

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
