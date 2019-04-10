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

	relChannelNode1Query = `MATCH (n:Node),(c:Channel)
	WHERE n.PubKey = {node1Pub} AND c.ChannelID = {channelID}
	CREATE (n)-[r:OPENED {
		Node1TimeLockDelta: {node1TimeLockDelta},
		Node1MinHtlc: {node1MinHtlc},
		Node1FeeBaseMsat: {node1FeeBaseMsat},
		Node1FeeRateMilliMsat: {node1FeeRateMilliMsat},
		Node1Disabled: {node1Disabled}
	} ]->(c)`

	relChannelNode2Query = `MATCH (n:Node),(c:Channel)
	WHERE n.PubKey = {node2Pub} AND c.ChannelID = {channelID}
	CREATE (n)-[r:OPENED {
		Node2TimeLockDelta: {node2TimeLockDelta},
		Node2MinHtlc: {node2MinHtlc},
		Node2FeeBaseMsat: {node2FeeBaseMsat},
		Node2FeeRateMilliMsat: {node2FeeRateMilliMsat},
		Node2Disabled: {node2Disabled}
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
			"channelID":             c.ChannelID,
			"node1Pub":              c.Node1Pub,
			"node1TimeLockDelta":    c.Node1Policy.TimeLockDelta,
			"node1MinHtlc":          c.Node1Policy.MinHtlc,
			"node1FeeBaseMsat":      c.Node1Policy.FeeBaseMsat,
			"node1FeeRateMilliMsat": c.Node1Policy.FeeRateMilliMsat,
			"node1Disabled":         c.Node1Policy.Disabled,
		}

		node2Values := map[string]interface{}{
			"channelID":             c.ChannelID,
			"node2Pub":              c.Node2Pub,
			"node2TimeLockDelta":    c.Node2Policy.TimeLockDelta,
			"node2MinHtlc":          c.Node2Policy.MinHtlc,
			"node2FeeBaseMsat":      c.Node2Policy.FeeBaseMsat,
			"node2FeeRateMilliMsat": c.Node2Policy.FeeRateMilliMsat,
			"node2Disabled":         c.Node2Policy.Disabled,
		}

		if _, err := ci.conn.ExecPipeline([]string{
			relChannelNode1Query,
			relChannelNode2Query,
		}, node1Values, node2Values); err != nil {
			return err
		}

		counter <- i
	}
	close(counter)
	return nil
}
