package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

func (e *Explorer) etherscanDeploymentBatch(ctx context.Context, chainID int64, contractAddresses []string) (*[]ent.Deployment, error) {
	const op = "app.getDeploymentBlockFromEtherscan()"
	<-e.tokens

	var baseURL, apikey string
	switch chainID {
	case ent.MAINNET, ent.HOLESKY, ent.SEPOLIA:
		baseURL, apikey = ent.EtherScanKeys[chainID], e.etherScanKey
	}

	var urlBuilder strings.Builder
	urlBuilder.Grow(170 + 42*len(contractAddresses))
	urlBuilder.WriteString(baseURL + "?module=contract&action=getcontractcreation&contractaddresses=")
	for _, address := range contractAddresses {
		urlBuilder.WriteString(address)
	}
	urlBuilder.WriteString("&apikey=" + apikey)

	resp, err := http.Get(urlBuilder.String())
	if err != nil {
		e.log.Error("Error making Etherscan API getContractCreation request", zap.Error(err))
		return &[]ent.Deployment{}, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		e.log.Error("Error reading Etherscan API getContractCreation response", zap.Error(err))
		return &[]ent.Deployment{}, fmt.Errorf("%s: %w", op, err)
	}

	var response ent.EtherScanResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		e.log.Error("Error decoding JSON from Etherscan API getContractCreation response", zap.Error(err))
		return &[]ent.Deployment{}, fmt.Errorf("%s: %w", op, err)
	}

	if len(response.Result) == 0 {
		e.log.Debug("No deployment transaction found")
		return &[]ent.Deployment{}, fmt.Errorf("%s: %w", op)
	}

	return &response.Result, nil
}

func (e *Explorer) etherscanDeployment(ctx context.Context, chainID int64, contractAddress string) (*ent.Deployment, error) {
	const op = "app.getDeploymentBlockFromEtherscan()"
	<-e.tokens

	var baseURL, apikey string
	switch chainID {
	case ent.MAINNET, ent.HOLESKY, ent.SEPOLIA:
		baseURL, apikey = ent.EtherScanKeys[chainID], e.etherScanKey
	}

	resp, err := http.Get(baseURL + "?module=contract&action=getcontractcreation&contractaddresses=" + contractAddress + "&apikey=" + apikey)
	if err != nil {
		e.log.Error("Error making Etherscan API getContractCreation request", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		e.log.Error("Error reading Etherscan API getContractCreation response", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var response ent.EtherScanResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		e.log.Error("Error decoding JSON from Etherscan API getContractCreation response", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(response.Result) == 0 {
		e.log.Debug("No deployment transaction found")
		return nil, fmt.Errorf("%s: %w", op)
	}

	return &response.Result[0], nil
}
