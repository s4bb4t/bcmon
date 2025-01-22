package app

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	i "github.com/s4bb4t/bcmon/internal/interfaces"
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
	sync.RWMutex
}

func NewSupervisor(producer i.Producer, storage i.Storage, graph i.Graph, log *slog.Logger, delay time.Duration) *Supervisor {
	ctx, cancel := context.WithCancel(context.Background())

	blocksCh := make(chan *types.Block, 1)

	return &Supervisor{
		producer: producer,
		storage:  storage,
		graph:    graph,

		usedContracts: make(map[string]struct{}),
		newContracts:  make(map[string]struct{}),

		blocksCh:    blocksCh,
		contractsCh: producer.Out(),
		errCh:       make(chan error),

		delay: delay,

		log:    log,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Spin starts main processes.
// 1. retrieve new contracts loop.
// 2. save new contracts loop.
// 3. updater loop.
// Because Spin have a loops inside, it will lock current goroutine - use .Stop() to stop the app
func (s *Supervisor) Spin() {
	s.produceBlocks().takeContracts()

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case addr := <-s.contractsCh:
				if _, exist := s.newContracts[addr]; !exist {
					if _, exist = s.usedContracts[addr]; !exist {
						s.newContracts[addr] = struct{}{}
					}
				}
			}
		}
	}()

	go func() {
		for err := range s.errCh {
			if err != nil {
				fmt.Println(err.Error())
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

func (s *Supervisor) InitContracts(init bool) error {
	for contract := range s.newContracts {
		s.log.Debug("Contract to init", slog.Any("new", contract))
		if init {
			if _, exist := s.graph.RealExist()[contract]; exist {
				delete(s.newContracts, contract)
				continue
			}
		} else {
			if err := s.storage.SaveContract(contract); err != nil {
				return err
			}
		}

		if err := s.graph.Init(contract); err != nil {
			return err
		}

		if err := s.graph.Create(contract); err != nil {
			return err
		}

		if err := s.graph.Deploy(contract); err != nil {
			return err
		}

		s.usedContracts[contract] = struct{}{}

		delete(s.newContracts, contract)
	}

	return nil
}

// LoadContracts loads already initialized graph's contracts
//
// LoadContracts use Must* function, so it may panic
func (s *Supervisor) LoadContracts() *Supervisor {
	s.graph.MustLoadContracts(s.newContracts)
	return s
}

func (s *Supervisor) takeContracts() *Supervisor {
	s.producer.Addresses(s.blocksCh)

	return s
}

func (s *Supervisor) produceBlocks() *Supervisor {
	ticker := time.NewTicker(3 * time.Second)

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				block := s.producer.Block()

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

func (s *Supervisor) Stop() {
	s.producer.Stop()
	s.cancel()

	close(s.contractsCh)
	close(s.blocksCh)
	close(s.errCh)
}

func (s *Supervisor) Reload() *Supervisor {

	return s
}
