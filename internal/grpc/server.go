package grpc

import (
	g "git.web3gate.ru/web3/nft/GraphForge/grpc/forge"
	"google.golang.org/grpc"
)

func InitForgeGRPC(deployer Deployer) *grpc.Server {
	s := grpc.NewServer()
	g.RegisterSubgraphServiceServer(s, nil)
	return s
}
