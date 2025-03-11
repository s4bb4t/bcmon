package interfaces

import (
	"context"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"math/big"
)

type (
	Producer interface {
		Produce(lastBlockNumber *big.Int) (chan *big.Int, chan *ent.Contract, chan error)
		Stop()
	}

	Graph interface {
		RealExist() map[string]struct{}
		Init(contract string) error
		Create(contract string) error
		Deploy(contract string) error
	}

	Storage interface {
		SaveContractForge(ctx context.Context, num *big.Int, contractID int64) error
		SaveBlock(ctx context.Context, num *big.Int) error
		BlockHandled(ctx context.Context, num *big.Int) error

		LastBlock() (*big.Int, error)
		Initialized(ctx context.Context, contract *ent.Contract) bool

		SaveContract(ctx context.Context, dep *ent.Contract) (contractID int64, err error)
	}

	Deployer interface {
		CreateSubgraph(context.Context, *ent.Contract) error
	}

	Detector interface {
		Type(ctx context.Context, deployment *ent.Contract) (string, error)
		LoadInfo(ctx context.Context, contract *ent.Contract) error
	}
)
