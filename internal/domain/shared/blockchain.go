package shared

// BlockchainNetwork represents supported blockchain networks.
type BlockchainNetwork string

const (
	NetworkTron     BlockchainNetwork = "tron"
	NetworkEthereum BlockchainNetwork = "ethereum"
	NetworkBitcoin  BlockchainNetwork = "bitcoin"
)

// String returns the string representation of the blockchain network.
func (n BlockchainNetwork) String() string {
	return string(n)
}

// IsValid returns true if the blockchain network is valid.
func (n BlockchainNetwork) IsValid() bool {
	switch n {
	case NetworkTron, NetworkEthereum, NetworkBitcoin:
		return true
	default:
		return false
	}
}
