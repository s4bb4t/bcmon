package grpc

import (
	"context"
	g "git.web3gate.ru/web3/nft/GraphForge/grpc/forge"
)

type deployerServer struct {
	g.UnimplementedSubgraphServiceServer
	deployer Deployer
}

func (s *deployerServer) CreateSubgraph(context.Context, *g.CreateSubgraphRequest) error {
	return nil
}
func (s *deployerServer) CreateSubgraphBatch(context.Context, string) error {
	return nil
}
func (s *deployerServer) DeleteSubgraph(context.Context, string) error {
	return nil
}
