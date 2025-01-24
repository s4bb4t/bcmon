package interfaces

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
)

type (
	Producer interface {
		Addresses(chan *types.Block)
		Block() *types.Block
		Out() chan string
		Stop()
	}

	Graph interface {
		RealExist() map[string]struct{}

		Init(contract string) error
		Create(contract string) error
		Deploy(contract string) error
	}

	Storage interface {
		LoadContracts(ctx context.Context, src, dest map[string]struct{})
		SaveContract(ctx context.Context, address string) error
		Initialized(ctx context.Context, dest map[string]struct{})
	}
)
