package main

import (
	"context"
	application "git.web3gate.ru/web3/nft/GraphForge/internal/app"
	"git.web3gate.ru/web3/nft/GraphForge/internal/config"
	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"git.web3gate.ru/web3/nft/GraphForge/internal/eth"
	"git.web3gate.ru/web3/nft/GraphForge/internal/graph"
	"git.web3gate.ru/web3/nft/GraphForge/internal/storage"
	appcloser "git.web3gate.ru/web3/nft/GraphForge/pkg/app_closer"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/logger"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/pgsql/migrator"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/pgsql/pgconnector"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func init() {
	cmd := exec.Command("npm", "install")
	cmd.Dir = "./"
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	log := logger.FromEnv("[graph-forge]")

	cfg := config.CreateConfig()

	closer := appcloser.InitCloser(nil)

	pgConnector, err := pgconnector.CreateConnection(ctx,
		cfg.Db.Postgres.GetDsn(),
		cfg.Db.Postgres.GetMaxOpenConns(),
		cfg.Db.Postgres.GetIdleConns(),
		cfg.Db.Postgres.GetIdleTime(),
		closer)
	if err != nil {
		log.Error("pgConnector creation error", zap.Any("err", err))
	}
	if err := migrator.Migrate(pgConnector); err != nil {
		log.Error("migrator error", zap.Any("err", err))
	}

	for _, network := range cfg.Networks {
		log := log.With(zap.String("network", network.Name))
		repo := storage.NewStorage(ctx, pgConnector, log)
		theGraph := graph.NewGraph(network.Name, cfg.GetSubgraphPath(), cfg.GetGraphNodeURL(), log)
		producer := eth.NewProducer(network.UpstreamURL, network.GetRequestDelay(), log, make(chan entity.Deployment))

		app := application.NewSupervisor(
			ctx,
			network.Name,
			producer,
			repo,
			theGraph,
			log,
			network.UpdateDelay,
		)

		closer.AddCloser(app.Stop, network.Name)

		go app.Spin()
	}

	select {
	case <-ctx.Done():
	}
}
