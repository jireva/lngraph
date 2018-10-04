package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/cheggaaa/pb"
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

	if err := json.Unmarshal(graphContent, &graph); err != nil {
		return err
	}

	fmt.Println("⚡ Importing nodes")
	bar := pb.New(len(graph.LightningNodes))
	bar.SetMaxWidth(80)
	bar.Start()
	for _, lnode := range graph.LightningNodes {
		err = lnode.create(conn)
		if err != nil {
			log.Fatal(err)
		}
		bar.Increment()
	}
	bar.Finish()

	if err := createLightningNodeIndexes(conn); err != nil {
		log.Fatal(err)
	}

	fmt.Println("⚡ Importing channels")
	bar = pb.New(len(graph.Channels))
	bar.SetMaxWidth(80)
	bar.Start()
	for _, channel := range graph.Channels {
		err = channel.create(conn)
		if err != nil {
			log.Fatal(err)
		}

		err = channel.createRelationships(conn)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}
	bar.Finish()

	if err := createChannelIndexes(conn); err != nil {
		log.Fatal(err)
	}

	return nil
}
