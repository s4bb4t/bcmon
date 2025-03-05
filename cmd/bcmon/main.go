package main

import (
	"context"
	application "git.web3gate.ru/web3/nft/GraphForge/internal/app"
	"git.web3gate.ru/web3/nft/GraphForge/internal/config"
	"git.web3gate.ru/web3/nft/GraphForge/internal/detector"
	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"git.web3gate.ru/web3/nft/GraphForge/internal/eth"
	"git.web3gate.ru/web3/nft/GraphForge/internal/graph"
	"git.web3gate.ru/web3/nft/GraphForge/internal/grpc"
	"git.web3gate.ru/web3/nft/GraphForge/internal/storage"
	appcloser "git.web3gate.ru/web3/nft/GraphForge/pkg/app_closer"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/logger"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/pgsql/migrator"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/pgsql/pgconnector"
	"github.com/ethereum/go-ethereum/ethclient"
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

	clients := make(map[string]*ethclient.Client)
	for _, network := range cfg.Networks {
		client, err := ethclient.Dial(network.UpstreamURL)
		if err != nil {
			log.Error("ethclient.Dial error", zap.Any("err", err))
			break
		}
		clients[network.Name] = client

		log := log.With(zap.String("network", network.Name))
		repo := storage.NewStorage(ctx, pgConnector, log)
		theGraph := graph.NewGraph(network.Name, cfg.GetSubgraphPath(), cfg.GetGraphNodeURL(), log)
		producer := eth.NewProducer(client, network.GetRequestDelay(), log, make(chan entity.Deployment))

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

	detect := detector.NewTokenDetector(clients, log)
	theGraph := graph.NewGraph("universal", cfg.GetSubgraphPath(), cfg.GetGraphNodeURL(), log)
	server := grpc.InitForgeGRPC(log, theGraph, detect)

	closer.AddCloser(server.GracefulStop, "grpc")

	<-ctx.Done()
	go func() {
		closer.CloseAll()
	}()
	stop()
	os.Exit(0)
}
