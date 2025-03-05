package detector

import (
	"context"
	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"strings"
)

const (
	ERC20Type   = "ERC20"
	ERC721Type  = "ERC721"
	ERC1155Type = "ERC1155"
	UnknownType = "Unknown"
)

type TokenDetector struct {
	clients map[string]*ethclient.Client
	log     *zap.Logger
}

func NewTokenDetector(clients map[string]*ethclient.Client, logger *zap.Logger) *TokenDetector {
	return &TokenDetector{
		clients: clients,
		log:     logger,
	}
}

func (td *TokenDetector) Type(ctx context.Context, deployment entity.Deployment) (string, error) {
	addr := common.HexToAddress(deployment.Contract)

	if td.isERC721(ctx, deployment.Network, addr) {
		td.log.Debug("Identified as ERC721", zap.String("contract", deployment.Contract))
		return ERC721Type, nil
	}

	if td.isERC1155(ctx, deployment.Network, addr) {
		td.log.Debug("Identified as ERC1155", zap.String("contract", deployment.Contract))
		return ERC1155Type, nil
	}

	if td.isERC20(ctx, deployment.Network, addr) {
		td.log.Debug("Identified as ERC20", zap.String("contract", deployment.Contract))
		return ERC20Type, nil
	}

	td.log.Warn("Could not determine token type", zap.String("contract", deployment.Contract))
	return UnknownType, nil
}

func (td *TokenDetector) isERC721(ctx context.Context, network string, contractAddress common.Address) bool {
	interfaceID := [4]byte{0x80, 0xac, 0x58, 0xcd}
	result, err := td.callSupportsInterface(ctx, network, contractAddress, interfaceID)
	if err != nil {
		td.log.Debug("Failed to check ERC721 support", zap.Error(err))
		return false
	}
	return result
}

func (td *TokenDetector) isERC1155(ctx context.Context, network string, contractAddress common.Address) bool {
	interfaceID := [4]byte{0xd9, 0xb6, 0x7a, 0x26}
	result, err := td.callSupportsInterface(ctx, network, contractAddress, interfaceID)
	if err != nil {
		td.log.Debug("Failed to check ERC1155 support", zap.Error(err))
		return false
	}
	return result
}

func (td *TokenDetector) isERC20(ctx context.Context, network string, contractAddress common.Address) bool {
	abiJSON := `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`

	Abi, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		td.log.Debug("Failed to parse ABI for balanceOf", zap.Error(err))
		return false
	}

	data, err := Abi.Pack("balanceOf", common.HexToAddress("0x0000000000000000000000000000000000000000"))
	if err != nil {
		td.log.Debug("Failed to pack data for balanceOf", zap.Error(err))
		return false
	}

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}
	result, err := td.clients[network].CallContract(ctx, msg, nil)
	if err != nil {
		td.log.Debug("Failed to call balanceOf", zap.Error(err))
		return false
	}

	return len(result) > 0
}

func (td *TokenDetector) callSupportsInterface(ctx context.Context, network string, contractAddress common.Address, interfaceID [4]byte) (bool, error) {
	abiJSON := `[{"constant":true,"inputs":[{"name":"interfaceId","type":"bytes4"}],"name":"supportsInterface","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"}]`

	Abi, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return false, err
	}

	data, err := Abi.Pack("supportsInterface", interfaceID)
	if err != nil {
		return false, err
	}

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}
	result, err := td.clients[network].CallContract(ctx, msg, nil)
	if err != nil {
		return false, err
	}

	var supported bool
	if err := Abi.UnpackIntoInterface(&supported, "supportsInterface", result); err != nil {
		return false, err
	}

	return supported, nil
}
