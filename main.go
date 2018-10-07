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
	chaintxnsFile := flag.String("chaintxns", "", "a file with the output of: lncli listchaintxns")
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

	if _, err := neo4j.CreateIndexes(conn); err != nil {
		log.Fatal(err)
	}

	if *graphFile != "" {
		if err := importGraph(conn, *graphFile); err != nil {
			log.Fatal(err)
		}
	}

	if *chaintxnsFile != "" {
		if err := importTransactions(conn, *chaintxnsFile); err != nil {
			log.Fatal(err)
		}
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

// importGraph reads graph data from a file and provides it to the neo4j importer.
func importGraph(conn bolt.Conn, graphFile string) error {
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

// importTransactions reads chain transactions data from a file and provides
// it to the neo4j importer.
func importTransactions(conn bolt.Conn, chaintxnsFile string) error {
	transactionsContent, err := ioutil.ReadFile(chaintxnsFile)
	if err != nil {
		return err
	}

	var txs ln.Transactions
	if err := json.Unmarshal(transactionsContent, &txs); err != nil {
		return err
	}

	fmt.Println("⚡ Importing chain transactions")
	bar := pb.New(len(txs.Transactions)).SetMaxWidth(80)
	bar.Start()
	c := make(chan int)
	go neo4j.ImportTransactions(conn, txs.Transactions, c)
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
