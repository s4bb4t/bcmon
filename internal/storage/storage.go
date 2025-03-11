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
