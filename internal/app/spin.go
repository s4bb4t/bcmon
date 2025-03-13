package app

import (
	"context"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"go.uber.org/zap"
)

// Spin starts the main processes of the Supervisor.
// It runs three main loops:
// 1. Retrieves new contracts from the producer.
// 2. Saves new contracts to storage.
// 3. Periodically initializes contracts in the graph.
// This function blocks the current goroutine and should be stopped using the Stop() method.
func (s *Supervisor) Spin() {
	blockNumber, err := s.storage.LastBlock(s.chainID)
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

				// костыль todo: убрать
				// ---------------------------------------------------------------------
				_type, err := s.explorer.Type(context.Background(), contract)
				if err != nil {
					s.log.Error("failed to get type of contract", zap.Error(err))
				}

				if _type == ent.ERC20Type {
					continue
				}
				// ---------------------------------------------------------------------
				//

				s.Lock()
				if _, exist := s.newContracts[contract.Address]; !exist {
					if _, exist = s.usedContracts[contract.Address]; !exist {
						s.newContracts[contract.Address] = struct{}{}
						s.contracts = append(s.contracts, contract)
						s.log.Debug("contract to initialize:", zap.String("net", contract.Network), zap.String("addr", contract.Address))
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
				blockID, err := s.storage.SaveBlock(context.Background(), block, s.chainID)
				if err != nil {
					errCh <- err
					continue
				}

				if err := s.InitContracts(blockID); err != nil {
					s.log.Error("init error", zap.Error(err))
					continue
				}

				if err := s.storage.BlockHandled(context.Background(), block, s.chainID); err != nil {
					errCh <- err
					continue
				}

				handled <- struct{}{}
			}
		}
	}()
}
