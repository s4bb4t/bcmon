package interfaces

import "github.com/ethereum/go-ethereum/core/types"

type (
	Producer interface {
		Addresses(chan *types.Block)
		Block() *types.Block
		Out() chan string
		Stop()
	}

	Graph interface {
		MustLoadContracts(dict map[string]struct{})
		RealExist() map[string]struct{}

		Init(contract string) error
		Create(contract string) error
		Deploy(contract string) error
	}

	Storage interface {
		SaveContract(address string) error
	}
)
