package main

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
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
	CREATE (n)-[r:c {
		Node1TimeLockDelta: {node1TimeLockDelta},
		Node1MinHtlc: {node1MinHtlc},
		Node1FeeBaseMsat: {node1FeeBaseMsat},
		Node1FeeRateMilliMsat: {node1FeeRateMilliMsat},
		Node1Disabled: {node1Disabled}
	} ]->(c)`

	relChannelNode2Query = `MATCH (n:Node),(c:Channel)
	WHERE n.PubKey = {node2Pub} AND c.ChannelID = {channelID}
	CREATE (n)-[r:c {
		Node2TimeLockDelta: {node2TimeLockDelta},
		Node2MinHtlc: {node2MinHtlc},
		Node2FeeBaseMsat: {node2FeeBaseMsat},
		Node2FeeRateMilliMsat: {node2FeeRateMilliMsat},
		Node2Disabled: {node2Disabled}
	} ]->(c)`

	indexChannelIDQuery = "CREATE INDEX ON :Channel(ChannelID)"
	indexCapacityQuery  = "CREATE INDEX ON :Channel(Capacity)"
)

// Channel represents a Lightning Network channel.
type Channel struct {
	ChannelID   uint64        `json:"channel_id,string"`
	ChanPoint   string        `json:"chan_point"`
	LastUpdate  uint32        `json:"last_update"`
	Node1Pub    string        `json:"node1_pub"`
	Node2Pub    string        `json:"node2_pub"`
	Capacity    int64         `json:"capacity,string"`
	Node1Policy RoutingPolicy `json:"node1_policy"`
	Node2Policy RoutingPolicy `json:"node2_policy"`
}

// RoutingPolicy represents a node policy in a lightning channel.
type RoutingPolicy struct {
	TimeLockDelta    uint32 `json:"time_lock_delta"`
	MinHtlc          int64  `json:"min_htlc,string"`
	FeeBaseMsat      int64  `json:"fee_base_msat,string"`
	FeeRateMilliMsat int64  `json:"fee_rate_milli_msat,string"`
	Disabled         bool   `json:"disabled"`
}

// create writes a lightning channel resource in neo4j.
func (c Channel) create(conn bolt.Conn) error {
	values := map[string]interface{}{
		"channelID":  c.ChannelID,
		"chainPoint": c.ChanPoint,
		"lastUpdate": c.LastUpdate,
		"capacity":   c.Capacity,
	}
	_, err := conn.ExecNeo(createChannelQuery, values)
	if err != nil {
		return err
	}

	return nil
}

// createChannelIndexes indexes lightning channels.
func createChannelIndexes(conn bolt.Conn) error {
	_, err := conn.ExecPipeline([]string{
		indexChannelIDQuery,
		indexCapacityQuery,
	}, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// createRelationships creates relationships between a lightning channel and
// its nodes.
func (c Channel) createRelationships(conn bolt.Conn) error {
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
	_, err := conn.ExecPipeline([]string{
		relChannelNode1Query,
		relChannelNode2Query,
	}, node1Values, node2Values)
	if err != nil {
		return err
	}

	return nil
}
