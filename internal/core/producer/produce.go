package producer

import (
	"context"
	"fmt"
	"git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"go.uber.org/zap"
	"math/big"
)

const (
	// ERC-721
	transfer = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

	// ERC-1155
	transferSingle = "0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62"
	transferBatch  = "0x4a39dc06d4c0dbc64b70af90fd698a233a518a4cb44f16935b4b89f1de659520"
)

func (p *Producer) Produce(lastBlockNumber *big.Int, handled chan struct{}) (chan *big.Int, chan *entity.Contract, chan error) {
	blocks := make(chan *big.Int)
	contracts := make(chan *entity.Contract)
	errCh := make(chan error)
	one := big.NewInt(1)

	go func() {
		blockNumber := lastBlockNumber
		for {
			select {
			case <-p.done:
				close(blocks)
				close(contracts)
				close(errCh)
				return
			case <-handled:
				block, err := p.client.BlockByNumber(context.Background(), blockNumber)
				if err != nil {
					errCh <- fmt.Errorf("failed to get block: %w", err)
					return
				}

				p.log.Debug("new block", zap.Int64("number", block.Number().Int64()))

				for _, tx := range block.Transactions() {
					receipt, err := p.client.TransactionReceipt(context.Background(), tx.Hash())
					if err != nil {
						p.log.Error("failed to get transaction receipt:", zap.Error(err))
						return
					}

					for _, logEntry := range receipt.Logs {
						if len(logEntry.Topics) < 1 {
							continue
						}

						switch logEntry.Topics[0].Hex() {
						case transferSingle, transferBatch:
							if p.excepted(logEntry.Address.String()) {
								continue
							}

							c := &entity.Contract{
								Network: p.network,
								ChainID: entity.Atoi[p.network],
								Address: logEntry.Address.String(),
								Type:    entity.ERC1155Type,
							}
							c.Found(block.Number())
							contracts <- c
						case transfer:
							if p.excepted(logEntry.Address.String()) {
								continue
							}

							c := &entity.Contract{
								Network: p.network,
								ChainID: entity.Atoi[p.network],
								Address: logEntry.Address.String(),
								Type:    entity.ERC721Type,
							}
							c.Found(block.Number())
							contracts <- c
						}
					}
				}
				blocks <- block.Number()
			}
			blockNumber.Add(blockNumber, one)
		}
	}()

	return blocks, contracts, errCh
}

func (p *Producer) Stop() {
	close(p.done)
}
