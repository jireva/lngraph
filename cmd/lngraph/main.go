package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/cheggaaa/pb"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/xsb/lngraph/ln"
	"github.com/xsb/lngraph/neo4j"
)

func main() {
	neo4jURL := flag.String("url", "bolt://localhost:7687", "neo4j url")
	graphFile := flag.String("graph", "", "a file with the output of: lncli describegraph")
	noDelete := flag.Bool("nodelete", false, "start without deleting all previous data")
	flag.Parse()

	conn, err := newNeo4jConnection(*neo4jURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if !*noDelete {
		fmt.Println("⚡ Deleting existing data")
		if err := neo4j.DeleteAll(conn); err != nil {
			log.Fatal(err)
		}
	}

	if err := neo4j.CreateIndexes(conn); err != nil {
		log.Fatal(err)
	}

	if err := importAll(conn, *graphFile); err != nil {
		log.Fatal(err)
	}
}

// newNeo4jConnection creates a bolt protocol connection to neo4j.
func newNeo4jConnection(url string) (bolt.Conn, error) {
	driver := bolt.NewDriver()
	conn, err := driver.OpenNeo(url)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// importAll reads source data files and provides them to all neo4j importers.
func importAll(conn bolt.Conn, graphFile string) error {
	if graphFile == "" {
		return nil
	}

	graphContent, err := ioutil.ReadFile(graphFile)
	if err != nil {
		return err
	}

	var graph struct {
		Nodes    ln.Nodes    `json:"nodes"`
		Channels ln.Channels `json:"edges"`
	}
	if err := json.Unmarshal(graphContent, &graph); err != nil {
		return err
	}

	fmt.Println("⚡ Importing nodes")
	bar := pb.New(len(graph.Nodes)).SetMaxWidth(80)
	bar.Start()
	c := make(chan int)
	go neo4j.ImportNodes(conn, graph.Nodes, c)
	for {
		_, ok := <-c
		if !ok {
			break
		}
		bar.Increment()
	}
	bar.Finish()

	fmt.Println("⚡ Importing channels")
	bar = pb.New(len(graph.Channels)).SetMaxWidth(80)
	bar.Start()
	c = make(chan int)
	go neo4j.ImportChannels(conn, graph.Channels, c)
	for {
		_, ok := <-c
		if !ok {
			break
		}
		bar.Increment()
	}
	bar.Finish()

	return nil
}
