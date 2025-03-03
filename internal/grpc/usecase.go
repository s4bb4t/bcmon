package grpc

import (
	"context"
)

type Deployer interface {
	CreateSubgraph(context.Context, string) error
	CreateSubgraphBatch(context.Context, string) error
	DeleteSubgraph(context.Context, string) error
}

//type Deployer interface {
//	CreateSubgraph(context.Context, *subgraph.CreateSubgraphRequest) (*subgraph.CreateSubgraphResponse, error)
//	CreateSubgraphBatch(context.Context, *subgraph.CreateSubgraphBatchRequest) (*subgraph.CreateSubgraphBatchResponse, error)
//	DeleteSubgraph(context.Context, *subgraph.DeleteSubgraphRequest) (*emptypb.Empty, error)
//}
