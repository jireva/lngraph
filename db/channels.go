package db

import (
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	createChannelQuery = `CREATE (c:Channel {
		ChannelID: $channelID,
		ChanPoint: $chainPoint,
		LastUpdate: $lastUpdate,
		Capacity: $capacity
	} )`

	relChannelNodeQuery = `MATCH (n:Node),(c:Channel)
	WHERE n.PubKey = $pub AND c.ChannelID = $channelID
	CREATE (n)-[r:OPENED {
		TimeLockDelta: $timeLockDelta,
		MinHtlc: $minHtlc,
		FeeBaseMsat: $feeBaseMsat,
		FeeRateMilliMsat: $feeRateMilliMsat,
		Disabled: $disabled
	} ]->(c)`
)

// ChannelsImporter implements a Neo4j importer for channels.
type ChannelsImporter struct {
	Driver neo4j.Driver
}

// Import gets multiple channel resources and imports them into Neo4j and
// creates relationships between each channel and its nodes.
func (ci ChannelsImporter) Import(channels []*lnrpc.ChannelEdge, counter chan int) error {
	session := ci.Driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()
	for i, c := range channels {
		channelValues := map[string]interface{}{
			"channelID":  c.ChannelId,
			"chainPoint": c.ChanPoint,
			"lastUpdate": time.Unix(int64(c.LastUpdate), 0).Format("2006-01-02 03:04"),
			"capacity":   c.Capacity,
		}
		if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			return tx.Run(createChannelQuery, channelValues)
		}); err != nil {
			return err
		}

		node1Values := map[string]interface{}{
			"channelID":        c.ChannelId,
			"pub":              c.Node1Pub,
			"timeLockDelta":    c.GetNode1Policy().GetTimeLockDelta(),
			"minHtlc":          c.GetNode1Policy().GetMinHtlc(),
			"feeBaseMsat":      c.GetNode1Policy().GetFeeBaseMsat(),
			"feeRateMilliMsat": c.GetNode1Policy().GetFeeRateMilliMsat(),
			"disabled":         c.GetNode1Policy().GetDisabled(),
		}

		node2Values := map[string]interface{}{
			"channelID":        c.ChannelId,
			"pub":              c.Node2Pub,
			"timeLockDelta":    c.GetNode2Policy().GetTimeLockDelta(),
			"minHtlc":          c.GetNode2Policy().GetMinHtlc(),
			"feeBaseMsat":      c.GetNode2Policy().GetFeeBaseMsat(),
			"feeRateMilliMsat": c.GetNode2Policy().GetFeeRateMilliMsat(),
			"disabled":         c.GetNode2Policy().GetDisabled(),
		}

		if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			return tx.Run(relChannelNodeQuery, node1Values)
		}); err != nil {
			return err
		}
		if _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			return tx.Run(relChannelNodeQuery, node2Values)
		}); err != nil {
			return err
		}

		counter <- i
	}
	close(counter)
	return nil
}
