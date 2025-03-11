package eth

//
//import (
//	"context"
//	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
//	"github.com/ethereum/go-ethereum"
//	"github.com/ethereum/go-ethereum/accounts/abi"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/core/types"
//	"github.com/ethereum/go-ethereum/ethclient"
//	"go.uber.org/zap"
//	"strings"
//	"time"
//)
//
//const (
//	// ERC-721
//	transfer = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
//
//	// ERC-1155
//	transferSingle = "0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62"
//	transferBatch  = "0x4a39dc06d4c0dbc64b70af90fd698a233a518a4cb44f16935b4b89f1de659520"
//)
//
//type producer struct {
//	protocol string
//	network  string
//
//	log    *zap.Logger
//	client *ethclient.Client
//
//	receiptsCh chan *types.Receipt
//	outCh      chan entity.Deployment
//
//	delay time.Duration
//
//	ctx    context.Context
//	cancel context.CancelFunc
//}
//
//func NewProducer(client *ethclient.Client, delay time.Duration, log *zap.Logger, contracts chan entity.Deployment) *producer {
//	receiptCh := make(chan *types.Receipt, 1)
//	ctx, cancel := context.WithCancel(context.Background())
//
//	producer := producer{client: client, receiptsCh: receiptCh, outCh: contracts, ctx: ctx, cancel: cancel, log: log, delay: delay}
//
//	producer.handleReceipts()
//
//	return &producer
//}
//
//func (p *producer) Addresses(in chan *types.Block) {
//	go func() {
//		for {
//			select {
//			case <-p.ctx.Done():
//				return
//			case block := <-in:
//				if block == nil {
//					p.log.Error("some problems with upstream node")
//					return
//				}
//
//				for _, tx := range block.Transactions() {
//					select {
//					case <-p.ctx.Done():
//						return
//					default:
//						receipt, err := p.client.TransactionReceipt(context.Background(), tx.Hash())
//						if err != nil {
//							p.log.Debug("Failed to get transaction receipt:", zap.Error(err))
//							continue
//						}
//
//						p.receiptsCh <- receipt
//						time.Sleep(p.delay)
//					}
//				}
//				p.log.Debug("Transactions receipts sent")
//			}
//		}
//	}()
//}
//
//func (p *producer) Block() (*types.Block, error) {
//	block, err := p.client.BlockByNumber(context.Background(), nil)
//	if err != nil {
//		p.log.Debug("Failed to get latest block:", zap.Error(err))
//		return nil, err
//	}
//
//	p.log.Debug("Got new block")
//
//	return block, nil
//}
//
//func (p *producer) handleReceipts() {
//	go func() {
//		for {
//			select {
//			case <-p.ctx.Done():
//				return
//			case receipt := <-p.receiptsCh:
//				for _, logEntry := range receipt.Logs {
//					if len(logEntry.Topics) < 1 {
//						continue
//					}
//
//					dep := entity.Deployment{
//						Protocol: p.protocol,
//						Network:  p.network,
//						Contract: logEntry.Address.Hex(),
//					}
//
//					switch logEntry.Topics[0].Hex() {
//					case transfer:
//						if p.isERC721(logEntry.Address) {
//							dep.Type = entity.ERC721Type
//							p.outCh <- dep
//
//						} else {
//							dep.Type = entity.ERC20Type
//							p.outCh <- dep
//						}
//					case transferSingle, transferBatch:
//						dep.Type = entity.ERC1155Type
//						p.outCh <- dep
//					}
//				}
//			}
//		}
//	}()
//}
//
//func (p *producer) isERC721(contractAddress common.Address) bool {
//	interfaceID := [4]byte{0x80, 0xac, 0x58, 0xcd}
//	result, err := p.callSupportsInterface(contractAddress, interfaceID)
//	if err != nil {
//		return false
//	}
//	return result
//}
//
//func (p *producer) callSupportsInterface(contractAddress common.Address, interfaceID [4]byte) (bool, error) {
//	Abi, err := abi.JSON(strings.NewReader(`[{"constant":true,"inputs":[{"name":"interfaceId","type":"bytes4"}],"name":"supportsInterface","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"}]`))
//	if err != nil {
//		return false, err
//	}
//
//	data, err := Abi.Pack("supportsInterface", interfaceID)
//	if err != nil {
//		return false, err
//	}
//
//	msg := ethereum.CallMsg{
//		To:   &contractAddress,
//		Data: data,
//	}
//	result, err := p.client.CallContract(context.Background(), msg, nil)
//	if err != nil {
//		return false, err
//	}
//
//	var supported bool
//	if err := Abi.UnpackIntoInterface(&supported, "supportsInterface", result); err != nil {
//		return false, err
//	}
//
//	return supported, nil
//}
//
//func (p *producer) Out() chan entity.Deployment {
//	return p.outCh
//}
//
//func (p *producer) Stop() {
//	p.cancel()
//}
