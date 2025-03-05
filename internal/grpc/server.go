package grpc

import (
	g "git.web3gate.ru/web3/nft/GraphForge/grpc/forge"
	"git.web3gate.ru/web3/nft/GraphForge/internal/interfaces"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func InitForgeGRPC(log *zap.Logger, deployer interfaces.Deployer, detector interfaces.Detector) *grpc.Server {
	s := grpc.NewServer()
	g.RegisterSubgraphServiceServer(s, &deployerServer{
		log: log,
		dep: deployer,
		dec: detector,
	})
	return s
}
