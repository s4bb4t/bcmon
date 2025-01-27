package migrator

import (
	"context"
	"fmt"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/pgsql/pgconnector"

	"github.com/pressly/goose"
)

func Migrate(connector pgconnector.ConnectionManager) (err error) {
	const op = "migrator.Migrate"

	conn, err := connector.GetConnection(context.Background(), pgconnector.DBReadWrite)
	if err != nil {
		return fmt.Errorf("%s: connector error: %w", op, err)
	}

	if err = goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("%s: dialect error: %w", op, err)
	}

	if err = goose.Up(conn.DB, "migrations"); err != nil {
		return fmt.Errorf("%s: migrate error: %w", op, err)
	}

	return err
}
