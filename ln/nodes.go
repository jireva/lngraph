package ln

// Nodes represents all nodes in the network.
type Nodes []Node

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
