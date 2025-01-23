package graph

import (
	"log/slog"
	"os"
	"os/exec"
)

type Graph struct {
	path    string
	network string

	log *slog.Logger
}

func NewGraph(network, path string, log *slog.Logger) *Graph {
	return &Graph{log: log, network: network, path: path}
}

func (g *Graph) RealExist() map[string]struct{} {
	files, err := os.ReadDir(g.path)
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

	cmd := exec.Command("graph", "init", contract, contract, "--from-contract", contract, "--network", g.network, "--skip-install", "--skip-git", "--abi", "../abi.json")
	cmd.Dir = g.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (g *Graph) Create(contract string) error {
	cmd := exec.Command("graph", "create", contract, "--node", "http://localhost:8020/")
	cmd.Dir = g.path + "/" + contract
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (g *Graph) Deploy(contract string) error {
	cmd := exec.Command("graph", "deploy", contract, "--node", "http://localhost:8020/", "--version-label", "v0.0.1")
	cmd.Dir = g.path + "/" + contract
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
