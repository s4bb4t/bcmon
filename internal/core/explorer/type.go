package explorer

import (
	"context"
	"fmt"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"strings"
)

func (e *Explorer) Type(ctx context.Context, deployment *ent.Contract) (string, error) {
	addr := common.HexToAddress(deployment.Address)

	if e.isERC721(ctx, deployment.Network, addr) {
		return ent.ERC721Type, nil
	}

	if e.isERC1155(ctx, deployment.Network, addr) {
		return ent.ERC1155Type, nil
	}

	if e.isERC20(ctx, deployment.Network, addr) {
		return ent.ERC20Type, nil
	}

	e.log.Warn("Could not determine token type", zap.String("contract", deployment.Address))
	return ent.UnknownType, nil
}

func (e *Explorer) isERC721(ctx context.Context, network string, contractAddress common.Address) bool {
	interfaceID := [4]byte{0x80, 0xac, 0x58, 0xcd}
	result, err := e.callSupportsInterface(ctx, network, contractAddress, interfaceID)
	if err != nil {
		e.log.Debug("Failed to check ERC721 support", zap.Error(err))
		return false
	}
	return result
}

func (e *Explorer) isERC1155(ctx context.Context, network string, contractAddress common.Address) bool {
	interfaceID := [4]byte{0xd9, 0xb6, 0x7a, 0x26}
	result, err := e.callSupportsInterface(ctx, network, contractAddress, interfaceID)
	if err != nil {
		e.log.Debug("Failed to check ERC1155 support", zap.Error(err))
		return false
	}
	return result
}

func (e *Explorer) isERC20(ctx context.Context, network string, contractAddress common.Address) bool {
	abiJSON := `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`

	Abi, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		e.log.Debug("Failed to parse ABI for balanceOf", zap.Error(err))
		return false
	}

	data, err := Abi.Pack("balanceOf", common.HexToAddress("0x0000000000000000000000000000000000000000"))
	if err != nil {
		e.log.Debug("Failed to pack data for balanceOf", zap.Error(err))
		return false
	}

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}
	result, err := e.clients[network].CallContract(ctx, msg, nil)
	if err != nil {
		e.log.Debug("Failed to call balanceOf", zap.Error(err))
		return false
	}

	return len(result) > 0
}

func (e *Explorer) callSupportsInterface(ctx context.Context, network string, contractAddress common.Address, interfaceID [4]byte) (bool, error) {
	abiJSON := `[{"constant":true,"inputs":[{"name":"interfaceId","type":"bytes4"}],"name":"supportsInterface","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"}]`

	Abi, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return false, fmt.Errorf("failed to parse ABI: %w", err)
	}

	data, err := Abi.Pack("supportsInterface", interfaceID)
	if err != nil {
		return false, fmt.Errorf("failed to pack data: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	result, err := e.clients[network].CallContract(ctx, msg, nil)
	if err != nil {
		if strings.Contains(err.Error(), "execution reverted") {
			return false, nil
		}
		return false, fmt.Errorf("call contract failed: %w", err)
	}

	var supported bool
	if err := Abi.UnpackIntoInterface(&supported, "supportsInterface", result); err != nil {
		return false, fmt.Errorf("failed to unpack result: %w", err)
	}

	return supported, nil
}
