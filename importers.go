package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/cheggaaa/pb"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

const (
	deleteAllQuery = "MATCH (n) DETACH DELETE n"
)

func deleteAll(conn bolt.Conn) error {
	fmt.Println("⚡ Deleting existing data")
	_, err := conn.ExecNeo(deleteAllQuery, nil)
	if err != nil {
		return err
	}
	return nil
}

func createIndexes(conn bolt.Conn) error {
	if err := createLightningNodeIndexes(conn); err != nil {
		log.Fatal(err)
	}

	if err := createChannelIndexes(conn); err != nil {
		log.Fatal(err)
	}
	return nil
}

func importGraph(conn bolt.Conn, graphFile string) error {
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

	return nil
}
