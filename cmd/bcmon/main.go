package main

import (
	"context"
	application "github.com/s4bb4t/bcmon/internal/app"
	"github.com/s4bb4t/bcmon/internal/config"
	"github.com/s4bb4t/bcmon/internal/eth"
	"github.com/s4bb4t/bcmon/internal/graph"
	"github.com/s4bb4t/bcmon/internal/storage"
	appcloser "github.com/s4bb4t/bcmon/pkg/app_closer"
	"github.com/s4bb4t/bcmon/pkg/pgsql/migrator"
	"github.com/s4bb4t/bcmon/pkg/pgsql/pgconnector"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func init() {
	cmd := exec.Command("npm", "install", "@graphprotocol/graph-cli")
	cmd.Dir = "./"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	cmd = exec.Command("npm", "install", "@graphprotocol/graph-ts")
	cmd.Dir = "./"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.CreateConfig()

	closer := appcloser.InitCloser(nil)

	pgConnector, err := pgconnector.CreateConnection(ctx,
		cfg.Db.Postgres.GetDsn(),
		cfg.Db.Postgres.GetMaxOpenConns(),
		cfg.Db.Postgres.GetIdleConns(),
		cfg.Db.Postgres.GetIdleTime(),
		closer)
	if err != nil {
		logger.Error("pgConnector creation error", slog.Any("err", err))
	}

	if err := migrator.Migrate(pgConnector); err != nil {
		logger.Error("migrator error", slog.Any("err", err))
	}

	repo := storage.NewStorage(ctx, pgConnector, logger)
	theGraph := graph.NewGraph(cfg.GetNetwork(), cfg.GetSubgraphPath(), logger)
	producer := eth.NewProducer(cfg.GetUpstreamURL(), cfg.GetRequestDelay(), logger)

	app := application.NewSupervisor(
		ctx,
		producer,
		repo,
		theGraph,
		logger,
		cfg.GetUpdateDelay(),
		cfg.GetInputData())

	if err := app.InitContracts(true); err != nil {
		panic(err)
	}

	//timer := time.NewTimer(30 * time.Second)

	go app.Spin()

	select {
	//case <-timer.C:
	//	app.Stop()
	case <-ctx.Done():
		app.Stop()
	}
}
