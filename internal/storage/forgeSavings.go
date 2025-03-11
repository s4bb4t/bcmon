package storage

import (
	"context"
	"fmt"
	"math/big"
)

func (s *storage) BlockHandled(ctx context.Context, num *big.Int) error {
	const op = "storage.BlockHandled"

	_, err := s.db.ExecContext(ctx, `update nft.forge_block set is_handled = true where block_number = $1`, num.Int64())
	if err != nil {
		return fmt.Errorf("%s: failed to insert: %w", op, err)
	}

	return err
}

func (s *storage) SaveBlock(ctx context.Context, num *big.Int) error {
	const op = "storage.SaveBlock"

	_, err := s.db.ExecContext(ctx, `INSERT INTO nft.forge_block (block_number) values($1) on conflict do nothing`, num.Int64())
	if err != nil {
		return fmt.Errorf("%s: failed to insert: %w", op, err)
	}

	return err
}

func (s *storage) SaveContractForge(ctx context.Context, num *big.Int, contractID int64) error {
	const op = "storage.SaveContract"

	_, err := s.db.ExecContext(ctx, `INSERT INTO nft.forge_deployment (forge_block_id, contract_id) values($1, $2)`, num.Int64(), contractID)
	if err != nil {
		return fmt.Errorf("%s: failed to insert: %w", op, err)
	}

	return err
}
