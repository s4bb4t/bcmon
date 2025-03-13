package app

import (
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	i "git.web3gate.ru/web3/nft/GraphForge/internal/interfaces"
	"go.uber.org/zap"
	"sync"
)

type Supervisor struct {
	explorer i.Detector
	producer i.Producer
	storage  i.Storage
	graph    i.Graph

	contracts []*ent.Contract
	chainID   int64

	usedContracts map[string]struct{}
	newContracts  map[string]struct{}

	log *zap.Logger

	sync.Mutex
}

// NewSupervisor initializes a new Supervisor instance.
// It sets up the context, producer, storage, graph, and other necessary components.
// It also loads existing contracts from storage and prepares channels for communication.
func NewSupervisor(
	explorer i.Detector,
	producer i.Producer,
	storage i.Storage,
	graph i.Graph,
	log *zap.Logger,
	chainId int64,
) *Supervisor {
	return &Supervisor{
		explorer: explorer,
		producer: producer,
		storage:  storage,
		graph:    graph,

		chainID: chainId,

		usedContracts: make(map[string]struct{}),
		newContracts:  make(map[string]struct{}),

		log: log,
	}
}
