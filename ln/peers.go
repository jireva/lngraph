package ln

// Peers represents a list of peer nodes.
type Peers struct {
	Peers []Peer `json:"peers"`
}

// Peer represents a peer node.
type Peer struct {
	PubKey    string `json:"pub_key"`
	Address   string `json:"address"`
	BytesSent uint64 `json:"bytes_sent,string"`
	BytesRecv uint64 `json:"bytes_recv,string"`
	SatSent   int64  `json:"sat_sent,string"`
	SatRecv   int64  `json:"sat_recv,string"`
	Inbound   bool   `json:"inbound"`
	PingTime  int64  `json:"ping_time,string"`
}
