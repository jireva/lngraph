package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

var (
	neo4jURL = "bolt://localhost:7687"
)

func newNeo4jConnection(url string) (bolt.Conn, error) {
	driver := bolt.NewDriver()
	conn, err := driver.OpenNeo(url)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func main() {
	// read from STDIN
	buf := new(bytes.Buffer)
	buf.ReadFrom(bufio.NewReader(os.Stdin))

	// json unmarshal
	var graph struct {
		LightningNodes []LightningNode `json:"nodes"`
		Channels       []Channel       `json:"edges"`
	}
	err := json.Unmarshal(buf.Bytes(), &graph)
	if err != nil {
		log.Fatal(err)
	}

	// connect to neo4j
	conn, err := newNeo4jConnection(neo4jURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// create and index lightning nodes
	for _, lnode := range graph.LightningNodes {
		err = lnode.create(conn)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = createLightningNodeIndexes(conn)
	if err != nil {
		log.Fatal(err)
	}

	// create channels
	for _, channel := range graph.Channels {
		err = channel.create(conn)
		if err != nil {
			log.Fatal(err)
		}
	}

	// create channel indexes
	err = createChannelIndexes(conn)
	if err != nil {
		log.Fatal(err)
	}

	// create channel relationships
	for _, channel := range graph.Channels {
		err = channel.createRelationships(conn)
		if err != nil {
			log.Fatal(err)
		}
	}
}
