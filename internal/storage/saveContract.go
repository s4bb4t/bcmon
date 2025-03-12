package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
)

func (s *storage) SaveContract(ctx context.Context, dep *ent.Contract) (int64, error) {
	const op = "storage.SaveContract"

	var depID int64
	var contractID int64
	query := `select id from nft.contract where chain_id = $1 and address = $2`
	if err := s.db.QueryRowContext(ctx, query, dep.ChainID, dep.Address).Scan(&contractID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			query := `INSERT INTO nft.deployment (block_number, deployer_address, contract_factory, tx_hash, timestamp, creation_byte_code) values($1, $2, $3, $4, $5, $6) RETURNING id`
			if err := s.db.QueryRowContext(ctx, query, dep.Deployment.BlockNumber, dep.Deployment.ContractCreator, dep.Deployment.ContractFactory, dep.Deployment.TxHash, dep.Deployment.Timestamp.String(), dep.Deployment.CreationByteCode).Scan(&depID); err != nil {
				return 0, fmt.Errorf("%s: failed to insert deoloyments: %w", op, err)
			}

			query = `INSERT INTO nft.contract (address, chain_id, deployment_id, type) values($1, $2, $3, $4) returning id`
			if err := s.db.QueryRowContext(ctx, query, dep.Address, dep.ChainID, depID, dep.Type).Scan(&contractID); err != nil {
				return 0, fmt.Errorf("%s: failed to insert contract: %w", op, err)
			}
			return contractID, nil
		}
		return 0, fmt.Errorf("%s: failed to get saved id: %w", op, err)
	}
	return contractID, nil
}
