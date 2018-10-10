package ln

// Node represents a lightning node.
type Node struct {
	LastUpdate uint32 `json:"last_update"`
	PubKey     string `json:"pub_key"`
	Alias      string `json:"alias"`
	Addresses  []struct {
		Network string `json:"network,omitempty"`
		Addr    string `json:"addr,omitempty"`
	} `json:"addresses,omitempty"`
	Color string `json:"color"`
}

// NodesImporter loads and imports node elements.
type NodesImporter interface {
	Import(nodes []Node, counter chan int) error
}

// NodesHandler handles and imports network nodes.
type NodesHandler struct {
	Nodes    []Node
	Importer NodesImporter
}

// NewNodesHandler creates a new NodesHandler.
func NewNodesHandler(ni NodesImporter) NodesHandler {
	return NodesHandler{
		Importer: ni,
	}
}

// Load loads nodes into a NodesHandler.
func (nh *NodesHandler) Load(nodes []Node) {
	nh.Nodes = nodes
}

// Import uses the specific importer method for importing data for persistence.
func (nh NodesHandler) Import(counter chan int) error {
	return nh.Importer.Import(nh.Nodes, counter)
}
