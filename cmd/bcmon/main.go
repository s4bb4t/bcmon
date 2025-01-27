package main

import (
	"context"
	application "git.web3gate.ru/web3/nft/GraphForge/internal/app"
	"git.web3gate.ru/web3/nft/GraphForge/internal/config"
	"git.web3gate.ru/web3/nft/GraphForge/internal/eth"
	"git.web3gate.ru/web3/nft/GraphForge/internal/graph"
	"git.web3gate.ru/web3/nft/GraphForge/internal/storage"
	appcloser "git.web3gate.ru/web3/nft/GraphForge/pkg/app_closer"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/pgsql/migrator"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/pgsql/pgconnector"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := cfg.Sepolia.ValidateNetwork(); err != nil {
			logger.Error("Sepolia", slog.Any("error", err))
			return
		}

		repo := storage.NewStorage(ctx, pgConnector, logger)
		theGraph := graph.NewGraph(cfg.Sepolia.GetNetwork(), cfg.GetSubgraphPath(), cfg.Sepolia.GetGraphNodeURL(), logger)
		producer := eth.NewProducer(cfg.Sepolia.GetUpstreamURL(), cfg.Sepolia.GetRequestDelay(), logger)

		app := application.NewSupervisor(
			ctx,
			producer,
			repo,
			theGraph,
			logger,
			cfg.Sepolia.GetUpdateDelay(),
			cfg.GetInputData())

		if err := app.InitContracts(true); err != nil {
			panic(err)
		}

		go app.Spin()
		select {
		case <-ctx.Done():
			app.Stop()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := cfg.Mainnet.ValidateNetwork(); err != nil {
			logger.Error("Mainnet", slog.Any("error", err))
			return
		}

		repo := storage.NewStorage(ctx, pgConnector, logger)
		theGraph := graph.NewGraph(cfg.Mainnet.GetNetwork(), cfg.GetSubgraphPath(), cfg.Mainnet.GetGraphNodeURL(), logger)
		producer := eth.NewProducer(cfg.Mainnet.GetUpstreamURL(), cfg.Mainnet.GetRequestDelay(), logger)

		app := application.NewSupervisor(
			ctx,
			producer,
			repo,
			theGraph,
			logger,
			cfg.Mainnet.GetUpdateDelay(),
			cfg.GetInputData())

		if err := app.InitContracts(true); err != nil {
			panic(err)
		}

		go app.Spin()
		select {
		case <-ctx.Done():
			app.Stop()
			wg.Done()
		}
	}()

	wg.Wait()
}
