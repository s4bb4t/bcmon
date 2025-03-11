package app

import (
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	i "git.web3gate.ru/web3/nft/GraphForge/internal/interfaces"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Supervisor struct {
	explorer i.Detector
	producer i.Producer
	storage  i.Storage
	graph    i.Graph

	usedContracts map[ent.Contract]struct{}
	newContracts  map[ent.Contract]struct{}

	delay time.Duration

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
	delay time.Duration,
) *Supervisor {
	return &Supervisor{
		explorer: explorer,
		producer: producer,
		storage:  storage,
		graph:    graph,

		usedContracts: make(map[ent.Contract]struct{}),
		newContracts:  make(map[ent.Contract]struct{}),

		delay: delay,

		log: log,
	}
}
