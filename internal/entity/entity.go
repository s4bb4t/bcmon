package entity

type Deployment struct {
	Protocol string
	Network  string
	Contract string
	Type     string
}

const (
	ERC20Type   = "ERC20"
	ERC721Type  = "ERC721"
	ERC1155Type = "ERC1155"
)
