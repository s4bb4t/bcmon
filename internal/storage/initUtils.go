package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"math/big"
)

func (s *storage) LastBlock() (*big.Int, error) {
	const op = "storage.LastBlock"

	var isHandled bool
	var blockNum int64
	if err := s.db.QueryRowContext(context.Background(), `select block_number, is_handled from nft.forge_block order by date desc limit 1`).Scan(&blockNum, &isHandled); err != nil {
		return nil, fmt.Errorf("%s: failed to insert: %w", op, err)
	}

	if isHandled {
		blockNum++
	}

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
