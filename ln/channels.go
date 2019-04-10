package ln

// Channel represents a lightning channel.
type Channel struct {
	ChannelID   uint64        `json:"channel_id,string"`
	ChanPoint   string        `json:"chan_point"`
	LastUpdate  int64         `json:"last_update"`
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

// ChannelsImporter loads and imports channel elements.
type ChannelsImporter interface {
	Import(channels []Channel, counter chan int) error
}

// ChannelsHandler handles and imports lightning channels.
type ChannelsHandler struct {
	Channels []Channel
	Importer ChannelsImporter
}

// NewChannelsHandler creates a new ChannelsHandler.
func NewChannelsHandler(ci ChannelsImporter) ChannelsHandler {
	return ChannelsHandler{
		Importer: ci,
	}
}

// Load loads channels into a ChannelsHandler.
func (ch *ChannelsHandler) Load(channels []Channel) {
	ch.Channels = channels
}

// Import uses the specific importer method for importing data for persistence.
func (ch ChannelsHandler) Import(counter chan int) error {
	return ch.Importer.Import(ch.Channels, counter)
}
