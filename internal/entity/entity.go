package entity

import (
	"fmt"
	"math/big"
	"time"
)

type (
	Deployment struct {
		ContractCreator  string `json:"contractCreator"`
		TxHash           string `json:"txHash"`
		BlockNumber      string `json:"blockNumber"`
		TimeUnix         string `json:"timestamp"`
		Timestamp        time.Time
		ContractFactory  string `json:"contractFactory"`
		CreationByteCode string `json:"creationBytecode"`
	}

	Contract struct {
		blockFoundAt *big.Int

		Network    string
		ChainID    int64
		Address    string
		Type       string
		Deployment *Deployment
	}

	AppBlock struct {
		ID        int
		Number    *big.Int
		IsHandled bool
	}

	AppDeployment struct {
		AppBlockID int
		ContractID int
	}

	EtherScanResponse struct {
		Result []Deployment `json:"result"`
	}
)

const (
	ERC20Type   = "ERC20"
	ERC721Type  = "ERC721"
	ERC1155Type = "ERC1155"
	UnknownType = "Unknown"

	MAINNET int64 = 1
	SEPOLIA int64 = 11155111
	HOLESKY int64 = 17000

	ETH_MAINNET_ADDRESS = "https://api.etherscan.io/api"
	ETH_SEPOLIA_ADDRESS = "https://api-sepolia.etherscan.io/api"
	ETH_HOLESKY_ADDRESS = "https://api-holesky.etherscan.io/api"
)

var (
	EtherScanKeys = map[int64]string{
		MAINNET: ETH_SEPOLIA_ADDRESS,
		SEPOLIA: ETH_SEPOLIA_ADDRESS,
		HOLESKY: ETH_HOLESKY_ADDRESS,
	}

	Atoi = map[string]int64{
		"mainnet": MAINNET,
		"sepolia": SEPOLIA,
		"holesky": HOLESKY,
	}

	Itoa = map[int64]string{
		MAINNET: "mainnet",
		SEPOLIA: "sepolia",
		HOLESKY: "holesky",
	}

	ErrNOTOK = fmt.Errorf("No data found")
)

func (c *Contract) FoundAt() *big.Int {
	return c.blockFoundAt
}

func (c *Contract) Found(num *big.Int) {
	c.blockFoundAt = num
}
