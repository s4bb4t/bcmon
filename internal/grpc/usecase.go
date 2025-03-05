package grpc

import (
	"context"
	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
)

type Deployer interface {
	CreateSubgraph(context.Context, entity.Deployment) error
	CreateSubgraphBatch(context.Context, entity.Deployment) error
	DeleteSubgraph(context.Context, entity.Deployment) error

	Type(context.Context, entity.Deployment) (string, error)
}

//type Deployer interface {
//	CreateSubgraph(context.Context, *subgraph.CreateSubgraphRequest) (*subgraph.CreateSubgraphResponse, error)
//	CreateSubgraphBatch(context.Context, *subgraph.CreateSubgraphBatchRequest) (*subgraph.CreateSubgraphBatchResponse, error)
//	DeleteSubgraph(context.Context, *subgraph.DeleteSubgraphRequest) (*emptypb.Empty, error)
//}
