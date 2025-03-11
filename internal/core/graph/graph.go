package graph

import (
	"context"
	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
)

type Graph struct {
	path    string
	network string
	nodeURL string

	log *zap.Logger
}

func NewGraph(network, path, node string, log *zap.Logger) *Graph {
	return &Graph{log: log, network: network, path: path, nodeURL: node}
}

func (g *Graph) RealExist() map[string]struct{} {
	files, err := os.ReadDir(g.path + "/" + g.network)
	if err != nil {
		return nil
	}

	filesMap := make(map[string]struct{})
	for _, file := range files {
		filesMap[file.Name()] = struct{}{}
	}

	return filesMap
}

func (g *Graph) Init(contract string) error {
	if err := os.Mkdir(g.path, 0644); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	cmd := exec.Command("graph", "init", g.network+"/"+contract, g.network+"/"+contract, "--from-contract", contract, "--network", g.network, "--skip-install", "--skip-git", "--abi", "../abi.json")
	cmd.Dir = g.path
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr

	g.log.Debug("graph-init")
	return cmd.Run()
}

func (g *Graph) Create(contract string) error {
	cmd := exec.Command("graph", "create", g.network+"/"+contract, "--node", g.nodeURL)
	cmd.Dir = g.path + "/" + g.network + "/" + contract

	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr

	g.log.Debug("graph-create")
	return cmd.Run()
}

func (g *Graph) Deploy(contract string) error {
	cmd := exec.Command("graph", "deploy", g.network+"/"+contract, "--node", g.nodeURL, "--version-label", "v0.0.1")
	cmd.Dir = g.path + "/" + g.network + "/" + contract
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	g.log.Debug("graph-deploy")
	return cmd.Run()
}

func (g *Graph) CreateSubgraph(ctx context.Context, contract *entity.Contract) error {
	if err := os.Mkdir(g.path, 0644); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	cmd := exec.Command("graph", "init", contract.Network+"/"+contract.Address, contract.Network+"/"+contract.Address, "--from-contract", contract.Address, "--network", contract.Network, "--skip-install", "--skip-git", "--abi", "../abi.json")
	cmd.Dir = g.path
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("graph", "create", contract.Network+"/"+contract.Address, "--node", g.nodeURL)
	cmd.Dir = g.path + "/" + g.network + "/" + contract.Address

	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("graph", "deploy", contract.Network+"/"+contract.Address, "--node", g.nodeURL, "--version-label", "v0.0.1")
	cmd.Dir = g.path + "/" + g.network + "/" + contract.Address
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
