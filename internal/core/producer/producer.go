package producer

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"sync"
)

type Producer struct {
	network string

	log        *zap.Logger
	client     *ethclient.Client
	exceptions map[string]struct{}

	done chan struct{}

	sync.RWMutex
}

func NewProducer(client *ethclient.Client, log *zap.Logger, network string) *Producer {
	return &Producer{client: client, log: log, done: make(chan struct{}), network: network, exceptions: make(map[string]struct{})}
}
