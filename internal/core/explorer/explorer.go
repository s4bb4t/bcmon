package explorer

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"time"
)

type Explorer struct {
	etherScanKey string
	clients      map[string]*ethclient.Client
	tokens       chan struct{}
	log          *zap.Logger
}

func NewTokenDetector(clients map[string]*ethclient.Client, logger *zap.Logger) *Explorer {
	tokens := make(chan struct{}, 5)
	go func() {
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ticker.C:
				for i := 0; i < 5; i++ {
					tokens <- struct{}{}
				}
			}
		}
	}()

	return &Explorer{
		clients:      clients,
		log:          logger,
		etherScanKey: "MR1U8E6ZVFY534W81WEQ7KUT6JAATTP9M1",
		tokens:       tokens,
	}
}
