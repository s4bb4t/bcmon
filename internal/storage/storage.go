package storage

import (
	"context"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/pgsql/pgconnector"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type storage struct {
	inputData map[string]struct{}

	db *sqlx.DB

	log *zap.Logger
}

func NewStorage(ctx context.Context, connector pgconnector.ConnectionManager, log *zap.Logger) *storage {
	db, err := connector.GetConnection(ctx, pgconnector.DBReadWrite)
	if err != nil {
		panic(err)
	}

	return &storage{log: log, db: db}
}

func (s *storage) SaveContract(ctx context.Context, address, network string) error {
	_, err := s.db.ExecContext(context.Background(), `INSERT INTO public.contract (address, network) values($1, $2)`, address, network)
	return err
}

func (s *storage) Initialized(ctx context.Context, network string, dest map[string]string) {
	rows, err := s.db.QueryContext(ctx, `select address from public.contract where network = $1`, network)
	if err != nil {
		panic(err)
	}

	var address string
	for rows.Next() {
		if err := rows.Scan(&address); err != nil {
			panic(err)
		}
		dest[address] = network
	}
}
