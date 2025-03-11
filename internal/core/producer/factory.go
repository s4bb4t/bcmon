package producer

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type Producer struct {
	network string

	log    *zap.Logger
	client *ethclient.Client

	done chan struct{}
}

func NewProducer(client *ethclient.Client, log *zap.Logger, network string) *Producer {
	return &Producer{client: client, log: log, done: make(chan struct{}), network: network}
}
