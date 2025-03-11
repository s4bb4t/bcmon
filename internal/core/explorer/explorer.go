package explorer

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type Explorer struct {
	etherScanKey string
	clients      map[string]*ethclient.Client
	tokens       chan struct{}
	log          *zap.Logger
}

func NewTokenDetector(clients map[string]*ethclient.Client, logger *zap.Logger) *Explorer {
	return &Explorer{
		clients: clients,
		log:     logger,
	}
}
