package pgconnector

import (
	"context"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type (
	// ConnectionType -
	ConnectionType int

	// ConnectionManager -
	ConnectionManager interface {
		GetConnection(context.Context, ConnectionType) (*sqlx.DB, error)
	}

	connectionManager struct {
		db *sqlx.DB
	}

	// Closer - исползуется при закрытии приложения
	Closer interface {
		AddCloser(func(), string)
	}
)

const (
	DBReadOnly  ConnectionType = iota << 1 // Читать лучше всегда с мастера и синка
	DBReadWrite                            // Записываем только на мастер
	DBReadAsync                            // async-нода используется как правило для аналитики
)

func CreateConnection(ctx context.Context, dsn string, maxOpenConns, maxIdleConns, connMaxIdleTime int, closer Closer) (ConnectionManager, error) {
	// this Pings the database trying to connect
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(time.Second * time.Duration(connMaxIdleTime))

	closer.AddCloser(func() {
		db.Close()
	}, "pgConnection")

	return &connectionManager{db: db}, nil
}

// Здесь будет логика по выдаче правильного коннекта для конфигурации master-slave-async
func (cm *connectionManager) GetConnection(ctx context.Context, connType ConnectionType) (*sqlx.DB, error) {
	return cm.db, nil
}
