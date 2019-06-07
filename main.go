package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/cheggaaa/pb"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/lightningnetwork/lnd/lncfg"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"github.com/xsb/lngraph/ln"
	"github.com/xsb/lngraph/neo4j"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	macaroon "gopkg.in/macaroon.v2"
)

func main() {
	// neo4j params
	neo4jURL := flag.String("url", "bolt://localhost:7687", "neo4j url")
	noDelete := flag.Bool("nodelete", false, "start without deleting all previous data")

	// gRPC params
	gRPCServer := flag.String("lnd-grpc", "127.0.0.1:10009", "LND gRPC server host:port")
	macaroonFile := flag.String("macaroon", "", "macaroon file")
	TLSCertFile := flag.String("tls-cert", "", "TLS certificate file")

	// files importing data
	graphFile := flag.String("graph", "", "a file with the output of: lncli describegraph")
	chaintxnsFile := flag.String("chaintxns", "", "a file with the output of: lncli listchaintxns")
	peersFile := flag.String("peers", "", "a file with the output of: lncli listpeers")
	flag.Parse()

	neo4jConn, err := neo4j.NewConnection(*neo4jURL)
	if err != nil {
		log.Fatal(err)
	}
	defer neo4jConn.Close()

	lndClient, err := newLNDClient(*gRPCServer, *macaroonFile, *TLSCertFile)
	if err != nil {
		log.Fatal(err)
	}

	if !*noDelete {
		fmt.Println("⚡ Deleting existing data")
		if err := neo4j.DeleteAll(neo4jConn); err != nil {
			log.Fatal(err)
		}
	}

	if _, err := neo4j.CreateIndexes(neo4jConn); err != nil {
		log.Fatal(err)
	}

	if *graphFile != "" {
		if err := importGraph(neo4jConn, *graphFile); err != nil {
			log.Fatal(err)
		}
	}

	if *chaintxnsFile != "" {
		if err := importTransactions(neo4jConn, *chaintxnsFile); err != nil {
			log.Fatal(err)
		}
	}

	if *peersFile != "" {
		if err := importPeers(neo4jConn, lndClient, *peersFile); err != nil {
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
func importPeers(conn bolt.Conn, lndClient lnrpc.LightningClient, peersFile string) error {
	ctx := context.Background()
	resp, err := lndClient.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return err
	}
	myPubKey := resp.GetIdentityPubkey()

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
	ph.Load(peers.Peers, myPubKey)
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

func newLNDClient(gRPCServer, macaroonFile, TLSCertFile string) (lnrpc.LightningClient, error) {
	var gRPCHost, gRPCPort string
	server := strings.Split(gRPCServer, ":")
	if len(server) == 2 {
		gRPCHost = server[0]
		gRPCPort = server[1]
	} else {
		return nil, errors.New("gRPC server must be provided in the format host:port")
	}

	providedMacaroon, err := ioutil.ReadFile(macaroonFile)
	if err != nil {
		return nil, err
	}

	grpcCredential, err := credentials.NewClientTLSFromFile(TLSCertFile, "")
	if err != nil {
		return nil, err
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(providedMacaroon); err != nil {
		return nil, err
	}
	macCredential := macaroons.NewMacaroonCredential(mac)

	// maxMsgRecvSize is the largest message our client will receive. We
	// set this to ~50Mb atm.
	//
	// Note: Got this config from lncli.
	maxMsgRecvSize := grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 50)

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", gRPCHost, gRPCPort),
		grpc.WithTransportCredentials(grpcCredential),
		grpc.WithPerRPCCredentials(macCredential),
		grpc.WithDialer(lncfg.ClientAddressDialer(fmt.Sprintf("%s", gRPCPort))),
		grpc.WithDefaultCallOptions(maxMsgRecvSize),
	)
	if err != nil {
		return nil, err
	}

	return lnrpc.NewLightningClient(conn), nil
}
