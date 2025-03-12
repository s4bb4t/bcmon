package app

import (
	"context"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"go.uber.org/zap"
)

// InitContracts initializes new contracts in the graph and saves them to storage.
// If `init` is true, it checks if the contract already exists in the graph before initializing.
// It also marks the contract as "used" after successful initialization.
func (s *Supervisor) InitContracts(blockNumber int64) error {
	ctx := context.Background()
	s.Lock()
	defer s.Unlock()

	for _, contract := range s.contracts {
		if err := s.explorer.LoadInfo(ctx, contract); err != nil {
			return err
		}

		if s.storage.Initialized(ctx, contract) {
			s.usedContracts[contract.Address] = struct{}{}
			continue
		}

		if err := s.graph.Init(contract.Address); err != nil {
			return err
		}
		if err := s.graph.Create(contract.Address); err != nil {
			return err
		}
		if err := s.graph.Deploy(contract.Address); err != nil {
			return err
		}

		contractID, err := s.storage.SaveContract(ctx, contract)
		if err != nil {
			return err
		}

		if err := s.storage.SaveContractForge(ctx, blockNumber, contractID); err != nil {
			return err
		}

		s.usedContracts[contract.Address] = struct{}{}
		delete(s.newContracts, contract.Address)

		s.log.Info("Deployed contract", zap.String("address", contract.Address))
	}

	s.contracts = []*ent.Contract{}
	return nil
}
