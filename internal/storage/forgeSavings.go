package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
)

func (s *storage) BlockHandled(ctx context.Context, num *big.Int, chainID int64) error {
	const op = "storage.BlockHandled"

	_, err := s.db.ExecContext(ctx, `update nft.forge_block set is_handled = true where block_number = $1 and chain_id = $2`, num.Int64(), chainID)
	if err != nil {
		return fmt.Errorf("%s: failed to insert: %w", op, err)
	}

	return err
}

func (s *storage) SaveBlock(ctx context.Context, num *big.Int, chainID int64) (int64, error) {
	const op = "storage.SaveBlock"

	var blockID int64
	if err := s.db.QueryRowContext(ctx, `select id from nft.forge_block where block_number = $1 and chain_id = $2`, num.Int64(), chainID).Scan(&blockID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := s.db.QueryRowContext(ctx, `INSERT INTO nft.forge_block (block_number, chain_id) values($1, $2) returning id`, num.Int64(), chainID).Scan(&blockID); err != nil {
				return 0, fmt.Errorf("%s: failed to insert: %w", op, err)
			}
			return blockID, nil
		}
		return 0, fmt.Errorf("%s: failed to check block: %w", op, err)
	}

	return blockID, nil
}

func (s *storage) SaveContractForge(ctx context.Context, num, contractID int64) error {
	const op = "storage.SaveContractForge"

	_, err := s.db.ExecContext(ctx, `INSERT INTO nft.forge_deployment (forge_block_id, contract_id) values($1, $2)`, num, contractID)
	if err != nil {
		return fmt.Errorf("%s: failed to insert: %w", op, err)
	}

	return err
}
