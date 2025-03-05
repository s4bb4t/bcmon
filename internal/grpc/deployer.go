package grpc

import (
	"context"
	"fmt"
	g "git.web3gate.ru/web3/nft/GraphForge/grpc/forge"
	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"git.web3gate.ru/web3/nft/GraphForge/internal/interfaces"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type deployerServer struct {
	log *zap.Logger
	dep interfaces.Deployer
	dec interfaces.Detector
	g.UnimplementedSubgraphServiceServer
}

func (s *deployerServer) CreateSubgraph(ctx context.Context, params *g.CreateSubgraphRequest) (*g.CreateSubgraphResponse, error) {
	deployment := entity.Deployment{
		Protocol: params.GetProtocol(),
		Network:  params.GetNetwork(),
		Contract: params.GetContractAddress(),
	}

	_type, err := s.dec.Type(ctx, deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to define type of contract: %w", err)
	}
	deployment.Type = _type

	if err := s.dep.CreateSubgraph(ctx, deployment); err != nil {
		return nil, fmt.Errorf("failed to create subgraph: %w", err)
	}

	return &g.CreateSubgraphResponse{SubgraphId: deployment.Network + "/" + deployment.Contract}, nil
}

func (s *deployerServer) CreateSubgraphBatch(ctx context.Context, params *g.CreateSubgraphBatchRequest) (*g.CreateSubgraphBatchResponse, error) {
	var ids []string

	for _, ent := range params.Subgraphs {
		deployment := entity.Deployment{
			Protocol: ent.GetProtocol(),
			Network:  ent.GetNetwork(),
			Contract: ent.GetContractAddress(),
		}

		_type, err := s.dec.Type(ctx, deployment)
		if err != nil {
			return nil, fmt.Errorf("failed to define type of contract: %w", err)
		}
		deployment.Type = _type

		if err := s.dep.CreateSubgraph(ctx, deployment); err != nil {
			return nil, fmt.Errorf("failed to create subgraph: %w", err)
		}

		ids = append(ids, deployment.Network+"/"+deployment.Contract)
	}

	return &g.CreateSubgraphBatchResponse{SubgraphIds: ids}, nil
}

func (s *deployerServer) DeleteSubgraph(ctx context.Context, params *g.DeleteSubgraphRequest) (*emptypb.Empty, error) {
	_, _ = ctx, params
	return nil, fmt.Errorf("temporary unimplemented ")
}
