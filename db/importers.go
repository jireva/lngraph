package db

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	deleteAllQuery = "MATCH (n) DETACH DELETE n"
)

// DeleteAll will remove existing neo4j nodes and relationships.
//
// The purpose is to allow new data to be entered without creating
// duplications, not to literally delete everything. For example, Indexes
// are not removed as they are useful in the future.
func DeleteAll(driver neo4j.Driver) error {
	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()
	result, err := session.Run(deleteAllQuery, nil)
	if err != nil {
		return err
	}
	_, err = result.Consume()
	if err != nil {
		return err
	}
	return nil
}
