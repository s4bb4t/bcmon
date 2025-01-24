package storage

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/s4bb4t/bcmon/pkg/pgsql/pgconnector"
	"log/slog"
)

type storage struct {
	inputData map[string]struct{}

	db *sqlx.DB

	log *slog.Logger
}

func NewStorage(ctx context.Context, connector pgconnector.ConnectionManager, log *slog.Logger) *storage {
	db, err := connector.GetConnection(ctx, pgconnector.DBReadWrite)
	if err != nil {
		panic(err)
	}

	return &storage{log: log, db: db}
}

func (s *storage) SaveContract(ctx context.Context, address string) error {
	_, err := s.db.ExecContext(context.Background(), `INSERT INTO public.contract (address) values($1)`, address)
	return err
}

func (s *storage) LoadContracts(ctx context.Context, src, dest map[string]struct{}) {
	dest = src
}

func (s *storage) Initialized(ctx context.Context, dest map[string]struct{}) {
	rows, err := s.db.QueryContext(ctx, `select * from public.contract`)
	if err != nil {
		panic(err)
	}

	var address string
	for rows.Next() {
		if err := rows.Scan(&address); err != nil {
			panic(err)
		}
		dest[address] = struct{}{}
	}
}
