package main

import (
	"flag"
	"log"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
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
		if err := deleteAll(conn); err != nil {
			log.Fatal(err)
		}
	}

	if err := createIndexes(conn); err != nil {
		log.Fatal(err)
	}

	if err := importGraph(conn, *graphFile); err != nil {
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
