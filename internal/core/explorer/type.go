package explorer

import (
	"context"
	"fmt"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/zap"
	"math/big"
	"strings"
	"time"
)

func (e *Explorer) Type(ctx context.Context, deployment *ent.Contract) (string, error) {
	addr := common.HexToAddress(deployment.Address)

	if e.isERC721(ctx, deployment.Network, addr) {
		return ent.ERC721Type, nil
	}
	if e.isERC1155(ctx, deployment.Network, addr) {
		return ent.ERC1155Type, nil
	}

	return ent.UnknownType, nil
}

func (e *Explorer) IsERC721(ctx context.Context, contract *ent.Contract) bool {
	return e.isERC721(ctx, contract.Network, common.HexToAddress(contract.Address))
}

func (e *Explorer) isERC721(ctx context.Context, network string, contractAddress common.Address) bool {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	interfaceID := [4]byte{0x80, 0xac, 0x58, 0xcd}
	result, err := e.callSupportsInterface(ctx, network, contractAddress, interfaceID)
	if err != nil {
		if !strings.Contains(err.Error(), "invalid opcode") && !strings.Contains(err.Error(), "invalid jump destination") && !strings.Contains(err.Error(), "unmarshal an empty string") {
			e.log.Debug("Failed to check ERC721 support", zap.Error(err))
		}
		return false
	}

	return result
}

func (e *Explorer) isERC1155(ctx context.Context, network string, contractAddress common.Address) bool {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	//interfaceID := [4]byte{0xd9, 0xb6, 0x7a, 0x26}
	//result, err := e.callSupportsInterface(ctx, network, contractAddress, interfaceID)
	//if err != nil {
	//	if !strings.Contains(err.Error(), "invalid opcode") && !strings.Contains(err.Error(), "invalid jump destination") {
	//		e.log.Warn("Failed to check ERC1155 support", zap.Error(err))
	//	}
	//	return false
	//}
	//
	//if result {
	//	return true
	//}

	block, err := e.clients[network].BlockNumber(ctx)
	if err != nil {
		e.log.Warn("failed to get last block", zap.Error(err))
		return false
	}

	transferSingleTopic := crypto.Keccak256Hash([]byte("TransferSingle(address,address,address,uint256,uint256)"))
	transferBatchTopic := crypto.Keccak256Hash([]byte("TransferBatch(address,address,address,uint256[],uint256[])"))

	q := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		Topics:    [][]common.Hash{{transferSingleTopic, transferBatchTopic}},
		FromBlock: new(big.Int).SetUint64(block - 50000),
		ToBlock:   new(big.Int).SetUint64(block),
	}
	fmt.Println(block, "-", block-50000)

	logs, err := e.clients[network].FilterLogs(ctx, q)
	fmt.Println(len(logs))
	if err != nil {
		fmt.Println(err)
		e.log.Warn("failed to get logs", zap.Error(err))
		return false
	}

	return len(logs) > 0
}

//func (e *Explorer) isERC20(ctx context.Context, network string, contractAddress common.Address) bool {
//
//	erc165ABI := `[{"constant":true,"inputs":[{"name":"interfaceId","type":"bytes4"}],"name":"supportsInterface","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"}]`
//	parsedABI, _ := abi.JSON(strings.NewReader(erc165ABI))
//	erc20InterfaceID := [4]byte{0x36, 0x37, 0x2b, 0x07}
//	data, _ := parsedABI.Pack("supportsInterface", erc20InterfaceID)
//	result, _ := e.clients[network].CallContract(ctx, ethereum.CallMsg{
//		To:   &contractAddress,
//		Data: data,
//	}, nil)
//
//
//	abiJSON := `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`
//
//	Abi, err := abi.JSON(strings.NewReader(abiJSON))
//	if err != nil {
//		e.log.Debug("Failed to parse ABI for balanceOf", zap.Error(err))
//		return false
//	}
//
//	data, err = Abi.Pack("balanceOf", common.HexToAddress("0x0000000000000000000000000000000000000000"))
//	if err != nil {
//		e.log.Debug("Failed to pack data for balanceOf", zap.Error(err))
//		return false
//	}
//
//	msg := ethereum.CallMsg{
//		To:   &contractAddress,
//		Data: data,
//	}
//	result, err = e.clients[network].CallContract(ctx, msg, nil)
//	if err != nil {
//		e.log.Debug("Failed to call balanceOf", zap.Error(err))
//		return false
//	}
//
//	return len(result) > 0
//}

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
		return false, fmt.Errorf("failed to unpack result: %w, for %s %s", err, network, contractAddress.String())
	}

	return supported, nil
}
