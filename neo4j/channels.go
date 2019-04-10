package neo4j

import (
	"time"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/xsb/lngraph/ln"
)

const (
	createChannelQuery = `CREATE (c:Channel {
		ChannelID: {channelID},
		ChanPoint: {chainPoint},
		LastUpdate: {lastUpdate},
		Capacity: {capacity}
	} )`

	relChannelNodeQuery = `MATCH (n:Node),(c:Channel)
	WHERE n.PubKey = {pub} AND c.ChannelID = {channelID}
	CREATE (n)-[r:OPENED {
		TimeLockDelta: {timeLockDelta},
		MinHtlc: {minHtlc},
		FeeBaseMsat: {feeBaseMsat},
		FeeRateMilliMsat: {feeRateMilliMsat},
		Disabled: {disabled}
	} ]->(c)`
)

// ChannelsImporter implements a Neo4j importer for channels.
type ChannelsImporter struct {
	conn bolt.Conn
}

// NewChannelsImporter creates a new ChannelsImporter.
func NewChannelsImporter(conn bolt.Conn) ChannelsImporter {
	return ChannelsImporter{
		conn: conn,
	}
}

// Import gets multiple channel resources and imports them into Neo4j and
// creates relationships between each channel and its nodes.
func (ci ChannelsImporter) Import(channels []ln.Channel, counter chan int) error {
	for i, c := range channels {
		channelValues := map[string]interface{}{
			"channelID":  c.ChannelID,
			"chainPoint": c.ChanPoint,
			"lastUpdate": time.Unix(c.LastUpdate, 0).Format("2006-01-02 03:04"),
			"capacity":   c.Capacity,
		}

		if _, err := ci.conn.ExecNeo(createChannelQuery, channelValues); err != nil {
			return err
		}

		node1Values := map[string]interface{}{
			"channelID":        c.ChannelID,
			"pub":              c.Node1Pub,
			"timeLockDelta":    c.Node1Policy.TimeLockDelta,
			"minHtlc":          c.Node1Policy.MinHtlc,
			"feeBaseMsat":      c.Node1Policy.FeeBaseMsat,
			"feeRateMilliMsat": c.Node1Policy.FeeRateMilliMsat,
			"disabled":         c.Node1Policy.Disabled,
		}

		node2Values := map[string]interface{}{
			"channelID":        c.ChannelID,
			"pub":              c.Node2Pub,
			"timeLockDelta":    c.Node2Policy.TimeLockDelta,
			"minHtlc":          c.Node2Policy.MinHtlc,
			"feeBaseMsat":      c.Node2Policy.FeeBaseMsat,
			"feeRateMilliMsat": c.Node2Policy.FeeRateMilliMsat,
			"disabled":         c.Node2Policy.Disabled,
		}

		if _, err := ci.conn.ExecPipeline([]string{
			relChannelNodeQuery,
			relChannelNodeQuery,
		}, node1Values, node2Values); err != nil {
			return err
		}

		counter <- i
	}
	close(counter)
	return nil
}
