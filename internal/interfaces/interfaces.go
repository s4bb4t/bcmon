package interfaces

import (
	"context"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"math/big"
)

type (
	Producer interface {
		Produce(lastBlockNumber *big.Int, handled chan struct{}) (chan *big.Int, chan *ent.Contract, chan error)
		Stop()
		Exception(contract string)
	}

	Graph interface {
		RealExist() map[string]struct{}
		Init(contract string) error
		Create(contract string) error
		Deploy(contract string) error
	}

	Storage interface {
		SaveContractForge(ctx context.Context, num, contractID int64) error
		SaveBlock(ctx context.Context, num *big.Int, chainID int64) (int64, error)
		BlockHandled(ctx context.Context, num *big.Int, chainID int64) error

		LastBlock(chainID int64) (*big.Int, error)
		Initialized(ctx context.Context, contract *ent.Contract) bool

		SaveContract(ctx context.Context, dep *ent.Contract) (contractID int64, err error)
	}

	Deployer interface {
		CreateSubgraph(context.Context, *ent.Contract) error
	}

	Detector interface {
		IsERC721(ctx context.Context, contract *ent.Contract) bool
		Type(ctx context.Context, contract *ent.Contract) (string, error)
		LoadInfo(ctx context.Context, contract *ent.Contract) error
	}
)
