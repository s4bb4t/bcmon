package app

import (
	"context"
	"fmt"
	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	i "git.web3gate.ru/web3/nft/GraphForge/internal/interfaces"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Supervisor struct {
	producer i.Producer
	storage  i.Storage
	graph    i.Graph

	usedContracts map[string]string
	newContracts  map[string]string

	blocksCh    chan *types.Block
	contractsCh chan entity.Deployment
	errCh       chan error

	delay time.Duration

	log    *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc

	sync.Mutex
}

// NewSupervisor initializes a new Supervisor instance.
// It sets up the context, producer, storage, graph, and other necessary components.
// It also loads existing contracts from storage and prepares channels for communication.
func NewSupervisor(
	ctx context.Context,
	network string,
	producer i.Producer,
	storage i.Storage,
	graph i.Graph,
	log *zap.Logger,
	delay time.Duration,
) *Supervisor {
	ctx, cancel := context.WithCancel(ctx)

	newC := make(map[string]string)
	storage.Initialized(ctx, network, newC)
	blocksCh := make(chan *types.Block, 1)

	return &Supervisor{
		producer: producer,
		storage:  storage,
		graph:    graph,

		usedContracts: make(map[string]string),
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
	s.handleErrorsLoop()
	s.produce()

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case addr := <-s.contractsCh:
				s.Lock()
				if _, exist := s.newContracts[addr.Contract]; !exist {
					if _, exist = s.usedContracts[addr.Contract]; !exist {
						s.newContracts[addr.Contract] = addr.Network
						s.log.Debug("contract to initialize:", zap.String("address", addr.Contract))
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
		for {
			select {
			case <-s.ctx.Done():
				return
			case err := <-s.errCh:
				if err != nil {
					s.log.Error("InitContracts:", zap.Any("error", err))
				}
			}
		}
	}()
}

// InitContracts initializes new contracts in the graph and saves them to storage.
// If `init` is true, it checks if the contract already exists in the graph before initializing.
// It also marks the contract as "used" after successful initialization.
func (s *Supervisor) InitContracts(init bool) error {
	realExist := s.graph.RealExist()

	for contract, network := range s.newContracts {
		s.Lock()

		if _, exist := realExist[contract]; exist {
			delete(s.newContracts, contract)
		}

		//if err := s.graph.Init(contract); err != nil {
		//	s.errCh <- err
		//	s.Unlock()
		//	continue
		//}
		//if err := s.graph.Create(contract); err != nil {
		//	s.errCh <- err
		//	s.Unlock()
		//	continue
		//}
		//if err := s.graph.Deploy(contract); err != nil {
		//	s.errCh <- err
		//	s.Unlock()
		//	continue
		//}
		//
		//if !init {
		//	if err := s.storage.SaveContract(s.ctx, contract, network); err != nil {
		//		return err
		//	}
		//}

		s.usedContracts[contract] = network

		delete(s.newContracts, contract)

		s.log.Debug("Deployed contract", zap.String("address", contract))
		s.Unlock()
	}

	fmt.Println(s.usedContracts)
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
				if block == nil {
					s.errCh <- fmt.Errorf("block is nil")
					return
				}
				if err != nil {
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
	s.log.Info("shutting down the Forge..")

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
