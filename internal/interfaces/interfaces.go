package interfaces

import (
	"context"
	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"github.com/ethereum/go-ethereum/core/types"
)

type (
	Producer interface {
		Addresses(chan *types.Block)
		Block() (*types.Block, error)
		Out() chan entity.Deployment
		Stop()
	}

	Graph interface {
		RealExist() map[string]struct{}
		Init(contract string) error
		Create(contract string) error
		Deploy(contract string) error
	}

	Storage interface {
		SaveContract(ctx context.Context, address, network string) error
		Initialized(ctx context.Context, network string, dest map[string]string)
	}
)
