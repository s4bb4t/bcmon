package main

import (
	"context"
	"fmt"
	application "git.web3gate.ru/web3/nft/GraphForge/internal/app"
	"git.web3gate.ru/web3/nft/GraphForge/internal/config"
	"git.web3gate.ru/web3/nft/GraphForge/internal/core/explorer"
	"git.web3gate.ru/web3/nft/GraphForge/internal/core/graph"
	"git.web3gate.ru/web3/nft/GraphForge/internal/core/producer"
	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"git.web3gate.ru/web3/nft/GraphForge/internal/grpc"
	"git.web3gate.ru/web3/nft/GraphForge/internal/storage"
	appcloser "git.web3gate.ru/web3/nft/GraphForge/pkg/app_closer"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/logger"
	"git.web3gate.ru/web3/nft/GraphForge/pkg/pgsql/pgconnector"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"io"
	"net"
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
		log.Panic("pgConnector creation error", zap.Any("err", err))
	}

	clients := make(map[string]*ethclient.Client)
	for _, network := range cfg.Networks {
		client, err := ethclient.Dial(network.UpstreamURL)
		if err != nil {
			log.Panic("ethclient.Dial error", zap.Any("err", err))
			break
		}
		clients[network.Name] = client

		log := log.With(zap.String("network", network.Name))
		repo := storage.NewStorage(ctx, pgConnector, log)
		theGraph := graph.NewGraph(network.Name, cfg.GetSubgraphPath(), cfg.GetGraphNodeURL(), log)
		prod := producer.NewProducer(client, log, network.Name)
		detect := explorer.NewTokenDetector(clients, log)

		app := application.NewSupervisor(
			detect,
			prod,
			repo,
			theGraph,
			log,
			entity.Atoi[network.Name],
		)

		closer.AddCloser(app.Stop, network.Name)

		//go app.Spin()
	}

	detect := explorer.NewTokenDetector(clients, log)
	theGraph := graph.NewGraph("universal", cfg.GetSubgraphPath(), cfg.GetGraphNodeURL(), log)
	repo := storage.NewStorage(ctx, pgConnector, log)
	server := grpc.InitForgeGRPC(log, theGraph, detect, repo)

	closer.AddCloser(server.GracefulStop, "grpc")

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GrpcPort()))
		if err != nil {
			log.Panic("failed to listen:", zap.Error(err))
		}

		log.Info(fmt.Sprintf("bc auth grpc server is running on %s", fmt.Sprintf(":%d", cfg.GrpcPort())))
		if err := server.Serve(lis); err != nil {
			log.Panic("failed to serve:", zap.Error(err))
		}
	}()

	<-ctx.Done()
	go func() {
		closer.CloseAll()
	}()
	stop()
	os.Exit(0)
}
