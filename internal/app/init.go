package app

import (
	"context"
	"go.uber.org/zap"
)

// InitContracts initializes new contracts in the graph and saves them to storage.
// If `init` is true, it checks if the contract already exists in the graph before initializing.
// It also marks the contract as "used" after successful initialization.
func (s *Supervisor) InitContracts(blockNumber int64) error {
	ctx := context.Background()

	for contract := range s.newContracts {
		if err := s.explorer.LoadInfo(ctx, &contract); err != nil {
			return err
		}

		if s.storage.Initialized(ctx, &contract) {
			s.usedContracts[contract] = struct{}{}
			continue
		}

		/*
			if err := s.graph.Init(contract.Address); err != nil {
				return err
			}
			if err := s.graph.Create(contract.Address); err != nil {
				return err
			}
			if err := s.graph.Deploy(contract.Address); err != nil {
				return err
			}
		*/

		contractID, err := s.storage.SaveContract(ctx, &contract)
		if err != nil {
			return err
		}

		if err := s.storage.SaveContractForge(ctx, blockNumber, contractID); err != nil {
			return err
		}

		s.usedContracts[contract] = struct{}{}
		delete(s.newContracts, contract)

		s.log.Debug("Deployed contract", zap.String("address", contract.Address))
	}

	s.log.Info("All new contracts successfully initialized!")
	return nil
}
