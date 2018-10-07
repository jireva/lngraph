package ln

// Channels represents all channels in the network.
type Channels []Channel

// Channel represents a lightning channel.
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
