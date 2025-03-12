package app

import (
	"context"
	"go.uber.org/zap"
)

// Spin starts the main processes of the Supervisor.
// It runs three main loops:
// 1. Retrieves new contracts from the producer.
// 2. Saves new contracts to storage.
// 3. Periodically initializes contracts in the graph.
// This function blocks the current goroutine and should be stopped using the Stop() method.
func (s *Supervisor) Spin() {
	blockNumber, err := s.storage.LastBlock()
	if err != nil {
		s.log.Error("failed to get last block", zap.Error(err))
	}

	done, handled := make(chan struct{}), make(chan struct{})
	blocks, contracts, errCh := s.producer.Produce(blockNumber, handled)
	handled <- struct{}{}

	go func() {
		for {
			select {
			case err := <-errCh:
				s.log.Error("producer error", zap.Error(err))
				close(done)
				return
			case <-done:
				return
			case contract := <-contracts:
				s.Lock()
				if _, exist := s.newContracts[*contract]; !exist {
					if _, exist = s.usedContracts[*contract]; !exist {
						s.newContracts[*contract] = struct{}{}
						s.log.Debug("contract to initialize:", zap.Any("ent", contract))
					}
				}
				s.Unlock()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case block := <-blocks:
				s.Lock()
				blockID, err := s.storage.SaveBlock(context.Background(), block)
				if err != nil {
					errCh <- err
					continue
				}

				if err := s.InitContracts(blockID); err != nil {
					errCh <- err
					continue
				}

				if err := s.storage.BlockHandled(context.Background(), block); err != nil {
					errCh <- err
					continue
				}

				handled <- struct{}{}
				s.Unlock()
			}
		}
	}()
}
