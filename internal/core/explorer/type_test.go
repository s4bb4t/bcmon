package explorer

import (
	"bytes"
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"testing"
)

func TestExplorer_TypeIs1155(t *testing.T) {
	client, _ := ethclient.Dial("https://b.dev.web3gate.ru:32443/5cb70a38-a49b-452a-ac94-ad5b98d5d482")
	ctx := context.Background()

	addr := common.HexToAddress("0xb97E8B386e783DFcE7Fa469D0E72b03C0f3B4A2A")
	//addr := common.HexToAddress("0xF1c35bdFdaF5B9A697095a996FEbe359071134e9")

	erc1155InterfaceID := [4]byte{0xd9, 0xb6, 0x7a, 0x26}
	callData := append([]byte{0x01, 0xff, 0xc9, 0xa7}, erc1155InterfaceID[:]...)

	msg := ethereum.CallMsg{
		To:   &addr,
		Data: callData,
	}

	result, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		t.Error(err)
	}

	if len(result) >= 32 && bytes.Equal(result[31:32], []byte{0x01}) {
		t.Skip()
	}

	transferSingleTopic := crypto.Keccak256Hash([]byte("TransferSingle(address,address,address,uint256,uint256)"))
	transferBatchTopic := crypto.Keccak256Hash([]byte("TransferBatch(address,address,address,uint256[],uint256[])"))

	q := ethereum.FilterQuery{
		Addresses: []common.Address{addr},
		Topics:    [][]common.Hash{{transferSingleTopic, transferBatchTopic}},
		FromBlock: new(big.Int).SetUint64(2000000),
		ToBlock:   new(big.Int).SetUint64(7894274),
	}

	logs, err := client.FilterLogs(ctx, q)
	if err != nil {
		t.Error(err)
	}

	if len(logs) > 0 {
		t.Skip()
	}

	t.Fail()
}

func TestExplorer_TypeIs721(t *testing.T) {
	client, _ := ethclient.Dial("https://b.dev.web3gate.ru:32443/5cb70a38-a49b-452a-ac94-ad5b98d5d482")
	ctx := context.Background()

	addr := common.HexToAddress("0x7B779c4751c6B848c647D8988941F28bC412357E")

	// 1. Проверка метода ownerOf (уникального для ERC-721)
	ownerOfSig := crypto.Keccak256([]byte("ownerOf(uint256)"))[:4]
	callDataOwnerOf := append(ownerOfSig, common.LeftPadBytes(common.Big0.Bytes(), 32)...) // Проверка с tokenId=0

	msgOwnerOf := ethereum.CallMsg{
		To:   &addr,
		Data: callDataOwnerOf,
	}

	_, err := client.CallContract(ctx, msgOwnerOf, nil)
	if err == nil {
		t.Skip("ownerOf method exists")
	}

	// 2. Проверка поддержки интерфейса ERC-721 через ERC-165
	erc721InterfaceID := [4]byte{0x80, 0xac, 0x58, 0xcd}
	callData := append([]byte{0x01, 0xff, 0xc9, 0xa7}, erc721InterfaceID[:]...)

	msg := ethereum.CallMsg{
		To:   &addr,
		Data: callData,
	}

	result, err := client.CallContract(ctx, msg, nil)
	if err == nil && len(result) >= 32 && bytes.Equal(result[31:32], []byte{0x01}) {
		t.Skip("ERC-721 interface confirmed")
	}

	// 3. Проверка события ApprovalForAll (уникального для ERC-721)
	approvalForAllTopic := crypto.Keccak256Hash([]byte("ApprovalForAll(address,address,bool)"))
	qApproval := ethereum.FilterQuery{
		Addresses: []common.Address{addr},
		Topics:    [][]common.Hash{{approvalForAllTopic}},
		FromBlock: new(big.Int).SetUint64(2000000),
		ToBlock:   new(big.Int).SetUint64(7000000),
	}

	logsApproval, err := client.FilterLogs(ctx, qApproval)
	if err != nil {
		t.Error(err)
	}
	if len(logsApproval) > 0 {
		t.Skip("Found ApprovalForAll events")
	}

	t.Fail()
}
