package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/lightningnetwork/lnd/lncfg"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/xsb/lngraph/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	macaroon "gopkg.in/macaroon.v2"
)

func main() {
	neo4jURL := flag.String("url", "bolt://localhost:7687", "neo4j url")
	noDelete := flag.Bool("nodelete", false, "start without deleting all previous data")
	gRPCServer := flag.String("lnd-grpc", "127.0.0.1:10009", "LND gRPC server host:port")
	macaroonFile := flag.String("macaroon", "", "macaroon file")
	TLSCertFile := flag.String("tls-cert", "", "TLS certificate file")
	flag.Parse()

	driver, err := neo4j.NewDriver(*neo4jURL, neo4j.NoAuth())
	if err != nil {
		log.Fatal(err)
	}
	defer driver.Close()

	lndClient, err := newLNDClient(*gRPCServer, *macaroonFile, *TLSCertFile)
	if err != nil {
		log.Fatal(err)
	}

	if !*noDelete {
		log.Println("Deleting existing data")
		if err := db.DeleteAll(driver); err != nil {
			log.Fatal(err)
		}
	}

	if err := db.CreateIndexes(driver); err != nil {
		log.Fatal(err)
	}

	if err := importGraph(driver, lndClient); err != nil {
		log.Fatal(err)
	}

	if err := importTransactions(driver, lndClient); err != nil {
		log.Fatal(err)
	}

	if err := importPeers(driver, lndClient); err != nil {
		log.Fatal(err)
	}
}

// importGraph reads graph data from lnd gRPC and provides it to the neo4j importer.
func importGraph(driver neo4j.Driver, lndClient lnrpc.LightningClient) error {
	ctx := context.Background()
	resp, err := lndClient.DescribeGraph(ctx, &lnrpc.ChannelGraphRequest{})
	if err != nil {
		return err
	}

	log.Println("Importing nodes")
	bar := pb.New(len(resp.GetNodes())).SetMaxWidth(80)
	bar.Start()
	ni := db.NodesImporter{driver}
	c := make(chan int)
	go ni.Import(resp.GetNodes(), c)
	for {
		_, ok := <-c
		if !ok {
			break
		}
		bar.Increment()
	}
	bar.Finish()

	log.Println("Importing channels")
	bar = pb.New(len(resp.GetEdges())).SetMaxWidth(80)
	bar.Start()
	ci := db.ChannelsImporter{driver}
	c = make(chan int)
	go ci.Import(resp.GetEdges(), c)
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

// importTransactions reads chain transactions data from lnd gRPC and provides
// it to the neo4j importer.
func importTransactions(driver neo4j.Driver, lndClient lnrpc.LightningClient) error {
	ctx := context.Background()
	resp, err := lndClient.GetTransactions(ctx, &lnrpc.GetTransactionsRequest{})
	if err != nil {
		return err
	}

	log.Println("Importing chain transactions")
	bar := pb.New(len(resp.GetTransactions())).SetMaxWidth(80)
	bar.Start()
	ti := db.TransactionsImporter{driver}
	c := make(chan int)
	go ti.Import(resp.GetTransactions(), c)
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

// importPeers reads peers data from lnd gRPC and provides it to the neo4j
// importer.
func importPeers(driver neo4j.Driver, lndClient lnrpc.LightningClient) error {
	ctx := context.Background()
	getInfoResp, err := lndClient.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return err
	}
	myPubKey := getInfoResp.GetIdentityPubkey()

	listPeersResp, err := lndClient.ListPeers(ctx, &lnrpc.ListPeersRequest{})
	if err != nil {
		return err
	}

	log.Println("Importing peers")
	bar := pb.New(len(listPeersResp.GetPeers())).SetMaxWidth(80)
	bar.Start()
	pi := db.PeersImporter{driver}
	c := make(chan int)
	go pi.Import(listPeersResp.GetPeers(), myPubKey, c)
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
