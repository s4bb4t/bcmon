package explorer

import (
	"context"
	"fmt"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
)

func (e *Explorer) LoadInfo(ctx context.Context, contract *ent.Contract) error {
	_type, err := e.Type(ctx, contract)
	if err != nil {
		return fmt.Errorf("failed to define type for %s, %s: %w", contract.Network, contract.Address, err)
	}

	if _type == ent.UnknownType {
		return fmt.Errorf("unknown type of contract: %s, %s", contract.Network, contract.Address)
	}

	deployments, err := e.etherscanDeployment(ctx, contract.ChainID, contract.Address)
	if err != nil {
		return fmt.Errorf("failed to get deployment for %s, %s: %w", contract.Network, contract.Address, err)
	}

	contract.Deployment = deployments
	contract.Type = _type

	return nil
}
