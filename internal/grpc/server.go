package grpc

import (
	g "git.web3gate.ru/web3/nft/GraphForge/grpc/forge"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func InitForgeGRPC(log *zap.Logger, deployer Deployer) *grpc.Server {
	s := grpc.NewServer()
	g.RegisterSubgraphServiceServer(s, &deployerServer{
		log: log,
		dep: deployer,
	})
	return s
}
