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
	noDelete := flag.Bool("nodelete", false, "start without deleting all previous data")
	graphFile := flag.String("graph", "", "a file with the output of: lncli describegraph")
	chaintxnsFile := flag.String("chaintxns", "", "a file with the output of: lncli listchaintxns")
	getInfoFile := flag.String("getinfo", "", "a file with the output of: lncli getinfo")
	peersFile := flag.String("peers", "", "a file with the output of: lncli listpeers")
	flag.Parse()

	conn, err := neo4j.NewConnection(*neo4jURL)
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

	if *peersFile != "" && *getInfoFile != "" {
		if err := importPeers(conn, *getInfoFile, *peersFile); err != nil {
			log.Fatal(err)
		}
	}
}

// importGraph reads graph data from a file and provides it to the neo4j importer.
func importGraph(conn bolt.Conn, graphFile string) error {
	graphContent, err := ioutil.ReadFile(graphFile)
	if err != nil {
		return err
	}

	var graph struct {
		Nodes    []ln.Node    `json:"nodes"`
		Channels []ln.Channel `json:"edges"`
	}
	if err := json.Unmarshal(graphContent, &graph); err != nil {
		return err
	}

	fmt.Println("⚡ Importing nodes")
	bar := pb.New(len(graph.Nodes)).SetMaxWidth(80)
	bar.Start()
	nh := ln.NewNodesHandler(neo4j.NewNodesImporter(conn))
	nh.Load(graph.Nodes)
	c := make(chan int)
	go nh.Import(c)
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
	ch := ln.NewChannelsHandler(neo4j.NewChannelsImporter(conn))
	ch.Load(graph.Channels)
	c = make(chan int)
	go ch.Import(c)
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

	var txs struct {
		Transactions []ln.Transaction `json:"transactions"`
	}
	if err := json.Unmarshal(transactionsContent, &txs); err != nil {
		return err
	}

	fmt.Println("⚡ Importing chain transactions")
	bar := pb.New(len(txs.Transactions)).SetMaxWidth(80)
	bar.Start()
	th := ln.NewTransactionsHandler(neo4j.NewTransactionsImporter(conn))
	th.Load(txs.Transactions)
	c := make(chan int)
	go th.Import(c)
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

// importPeers reads peers data from a file and provides it to the neo4j
// importer.
func importPeers(conn bolt.Conn, getInfoFile, peersFile string) error {
	getInfoContent, err := ioutil.ReadFile(getInfoFile)
	if err != nil {
		return err
	}

	var getInfo struct {
		IdentityPubkey string `json:"identity_pubkey"`
	}
	if err := json.Unmarshal(getInfoContent, &getInfo); err != nil {
		return err
	}

	peersContent, err := ioutil.ReadFile(peersFile)
	if err != nil {
		return err
	}

	var peers struct {
		Peers []ln.Peer `json:"peers"`
	}
	if err := json.Unmarshal(peersContent, &peers); err != nil {
		return err
	}

	fmt.Println("⚡ Importing peers")
	bar := pb.New(len(peers.Peers)).SetMaxWidth(80)
	bar.Start()
	ph := ln.NewPeersHandler(neo4j.NewPeersImporter(conn))
	ph.Load(peers.Peers, getInfo.IdentityPubkey)
	c := make(chan int)
	go ph.Import(c)
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
