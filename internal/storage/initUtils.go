package storage

import (
	"context"
	"fmt"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"math/big"
)

func (s *storage) LastBlock() (*big.Int, error) {
	const op = "storage.LastBlock"

	var blockNum int64
	if err := s.db.QueryRowContext(context.Background(), `select block_number from nft.forge_block order by date desc limit 1`).Scan(&blockNum); err != nil {
		return nil, fmt.Errorf("%s: failed to insert: %w", op, err)
	}

	return big.NewInt(blockNum), nil
}

func (s *storage) Initialized(ctx context.Context, contract *ent.Contract) bool {
	var initialized bool
	if err := s.db.QueryRowContext(ctx, `select 1 where chain_id = $1 and address = $2`, contract.ChainID, contract.Address).Scan(&initialized); err != nil {
		panic(err)
	}

	return initialized
}
