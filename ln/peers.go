package ln

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

// PeersImporter loads and imports peer node elements.
type PeersImporter interface {
	Import(peers []Peer, myPubKey string, counter chan int) error
}

// PeersHandler handles and imports peer nodes.
type PeersHandler struct {
	Peers    []Peer
	myPubKey string
	Importer PeersImporter
}

// NewPeersHandler creates a new PeersHandler.
func NewPeersHandler(pi PeersImporter) PeersHandler {
	return PeersHandler{
		Importer: pi,
	}
}

// Load loads peers into a PeersHandler.
func (ph *PeersHandler) Load(peers []Peer, myPubKey string) {
	ph.Peers = peers
	ph.myPubKey = myPubKey
}

// Import uses the specific importer method for importing data for persistence.
func (ph PeersHandler) Import(counter chan int) error {
	return ph.Importer.Import(ph.Peers, ph.myPubKey, counter)
}
