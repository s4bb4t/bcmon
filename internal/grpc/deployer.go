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
	log  *zap.Logger
	dep  interfaces.Deployer
	dec  interfaces.Detector
	repo interfaces.Storage
	g.UnimplementedSubgraphServiceServer
}

func (s *deployerServer) CreateSubgraph(ctx context.Context, params *g.CreateSubgraphRequest) (*g.CreateSubgraphResponse, error) {
	contract := &entity.Contract{
		Network: params.GetNetwork(),
		ChainID: entity.Atoi[params.GetNetwork()],
		Address: params.GetContractAddress(),
	}

	if s.repo.Initialized(ctx, contract) {
		return &g.CreateSubgraphResponse{SubgraphId: contract.Network + "/" + contract.Address}, nil
	}

	if err := s.dec.LoadInfo(ctx, contract); err != nil {
		return nil, fmt.Errorf("failed to load info about contract")
	}

	contractId, err := s.repo.SaveContract(ctx, contract)
	if err != nil {
		return nil, fmt.Errorf("failed to save contract")
	}

	if err := s.dep.CreateSubgraph(ctx, contract); err != nil {
		return nil, fmt.Errorf("failed to create subgraph: %w", err)
	}

	fmt.Println(0, contractId)

	if err := s.repo.SaveContractForge(ctx, 0, contractId); err != nil {
		return nil, fmt.Errorf("failed to save deployment")
	}

	s.log.Info("deployed new contract", zap.String("address", contract.Network+"/"+contract.Address))
	return &g.CreateSubgraphResponse{SubgraphId: contract.Network + "/" + contract.Address}, nil
}

func (s *deployerServer) CreateSubgraphBatch(ctx context.Context, params *g.CreateSubgraphBatchRequest) (*g.CreateSubgraphBatchResponse, error) {
	var ids []string

	for _, ent := range params.Subgraphs {
		contract := &entity.Contract{
			Network: ent.GetNetwork(),
			Address: ent.GetContractAddress(),
		}

		if s.repo.Initialized(ctx, contract) {
			return nil, nil
		}

		_type, err := s.dec.Type(ctx, contract)
		if err != nil {
			return nil, fmt.Errorf("failed to define type of contract: %w", err)
		}
		contract.Type = _type

		if err := s.dep.CreateSubgraph(ctx, contract); err != nil {
			return nil, fmt.Errorf("failed to create subgraph: %w", err)
		}

		ids = append(ids, contract.Network+"/"+contract.Address)
	}

	return &g.CreateSubgraphBatchResponse{SubgraphIds: ids}, nil
}

func (s *deployerServer) DeleteSubgraph(ctx context.Context, params *g.DeleteSubgraphRequest) (*emptypb.Empty, error) {
	_, _ = ctx, params
	return nil, fmt.Errorf("temporary unimplemented ")
}
