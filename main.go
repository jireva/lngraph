package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

func main() {
	neo4jURL := flag.String("url", "bolt://localhost:7687", "neo4j url")
	graphFile := flag.String("graph", "", "a file with the output of: lncli describegraph")
	flag.Parse()

	conn, err := newNeo4jConnection(*neo4jURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := importGraph(*graphFile, conn); err != nil {
		log.Fatal(err)
	}
}

func newNeo4jConnection(url string) (bolt.Conn, error) {
	driver := bolt.NewDriver()
	conn, err := driver.OpenNeo(url)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func importGraph(graphFile string, conn bolt.Conn) error {
	if graphFile == "" {
		return nil
	}

	graphContent, err := ioutil.ReadFile(graphFile)
	if err != nil {
		return err
	}

	var graph struct {
		LightningNodes []LightningNode `json:"nodes"`
		Channels       []Channel       `json:"edges"`
	}
	err = json.Unmarshal(graphContent, &graph)
	if err != nil {
		return err
	}

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

	return nil
}
