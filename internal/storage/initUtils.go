package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"math/big"
)

func (s *storage) LastBlock(chainID int64) (*big.Int, error) {
	const op = "storage.LastBlock"

	var blockNum int64
	if err := s.db.QueryRowContext(context.Background(), `select block_number from nft.forge_block where is_handled = true and chain_id = $1 order by date desc limit 1`, chainID).Scan(&blockNum); err != nil {
		return nil, fmt.Errorf("%s: failed to insert: %w", op, err)
	}

	blockNum++

	return big.NewInt(blockNum), nil
}

func (s *storage) Initialized(ctx context.Context, contract *ent.Contract) bool {
	var initialized bool
	query := `select 1 from nft.forge_deployment d join nft.contract c on d.contract_id = c.id where c.chain_id = $1 and c.address = $2`
	if err := s.db.QueryRowContext(ctx, query, contract.ChainID, contract.Address).Scan(&initialized); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		panic(err)
	}

	return initialized
}
