package app

import (
	"context"
	i "git.web3gate.ru/web3/nft/GraphForge/internal/interfaces"
	"github.com/ethereum/go-ethereum/core/types"
	"log/slog"
	"sync"
	"time"
)

type Supervisor struct {
	producer i.Producer
	storage  i.Storage
	graph    i.Graph

	usedContracts map[string]struct{}
	newContracts  map[string]struct{}

	blocksCh    chan *types.Block
	contractsCh chan string
	errCh       chan error

	delay time.Duration

	log    *slog.Logger
	ctx    context.Context
	cancel context.CancelFunc

	sync.Mutex
}

// NewSupervisor initializes a new Supervisor instance.
// It sets up the context, producer, storage, graph, and other necessary components.
// It also loads existing contracts from storage and prepares channels for communication.
func NewSupervisor(
	ctx context.Context,
	producer i.Producer,
	storage i.Storage,
	graph i.Graph,
	log *slog.Logger,
	delay time.Duration,
	inputData []string) *Supervisor {
	ctx, cancel := context.WithCancel(ctx)

	input := make(map[string]struct{})

	newC := make(map[string]struct{})
	if len(inputData) != 0 {
		for _, v := range inputData {
			input[v] = struct{}{}
		}
		storage.LoadContracts(ctx, input, newC)
	} else {
		storage.Initialized(ctx, newC)
	}

	blocksCh := make(chan *types.Block, 1)

	return &Supervisor{
		producer: producer,
		storage:  storage,
		graph:    graph,

		usedContracts: make(map[string]struct{}),
		newContracts:  newC,

		blocksCh:    blocksCh,
		contractsCh: producer.Out(),
		errCh:       make(chan error, 1),

		delay: delay,

		log:    log,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Spin starts the main processes of the Supervisor.
// It runs three main loops:
// 1. Retrieves new contracts from the producer.
// 2. Saves new contracts to storage.
// 3. Periodically initializes contracts in the graph.
// This function blocks the current goroutine and should be stopped using the Stop() method.
func (s *Supervisor) Spin() {
	s.log.Info("spinFunc: running app")

	s.handleErrorsLoop()
	s.produce()

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case addr := <-s.contractsCh:
				s.Lock()
				if _, exist := s.newContracts[addr]; !exist {
					if _, exist = s.usedContracts[addr]; !exist {
						s.newContracts[addr] = struct{}{}
						s.log.Debug("got new contract to initialize:", slog.String("address", addr))
					}
				}
				s.Unlock()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(s.delay)
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				select {
				case <-s.ctx.Done():
					return
				case s.errCh <- s.InitContracts(false):
				}
			}
		}
	}()
}

func (s *Supervisor) handleErrorsLoop() {
	go func() {
		for err := range s.errCh {
			if err != nil {
				s.log.Error("InitContracts:", slog.Any("error", err))
			}
		}
	}()
}

// InitContracts initializes new contracts in the graph and saves them to storage.
// If `init` is true, it checks if the contract already exists in the graph before initializing.
// It also marks the contract as "used" after successful initialization.
func (s *Supervisor) InitContracts(init bool) error {
	realExist := s.graph.RealExist()

	for contract := range s.newContracts {
		s.Lock()

		if _, exist := realExist[contract]; exist {
			delete(s.newContracts, contract)
		}

		if !init {
			if err := s.storage.SaveContract(s.ctx, contract); err != nil {
				return err
			}
		}

		if err := s.graph.Init(contract); err != nil {
			s.errCh <- err
		}
		if err := s.graph.Create(contract); err != nil {
			s.errCh <- err
		}
		if err := s.graph.Deploy(contract); err != nil {
			s.errCh <- err
		}

		s.usedContracts[contract] = struct{}{}

		delete(s.newContracts, contract)

		s.log.Debug("Deployed contract", slog.String("address", contract))
		s.Unlock()
	}

	s.log.Info("All new contracts successfully initialized!")
	return nil
}

// produce starts the process of fetching blocks from the producer and sending them to the blocks channel.
// It uses a ticker to periodically fetch blocks and ensures the process stops when the context is canceled.
func (s *Supervisor) produce() *Supervisor {
	s.producer.Addresses(s.blocksCh)

	ticker := time.NewTicker(s.delay / 2)

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				block, err := s.producer.Block()
				if err != nil || block == nil {
					s.errCh <- err
				}

				select {
				case <-s.ctx.Done():
					return
				case s.blocksCh <- block:
				}
			}
		}
	}()

	return s
}

// Stop gracefully shuts down the Supervisor.
// It stops the producer, cancels the context, and closes all communication channels.
func (s *Supervisor) Stop() {
	s.log.Info("shutting down the app")

	s.producer.Stop()
	s.cancel()

	close(s.contractsCh)
	close(s.blocksCh)
	close(s.errCh)
}

// Reload is a placeholder function for reloading or reinitializing the Supervisor.
// Currently, it does nothing but returns the Supervisor instance.
func (s *Supervisor) Reload() *Supervisor {

	return s
}
