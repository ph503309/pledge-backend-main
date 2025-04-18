package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"pledge-backend/api/models"
	"pledge-backend/config"
	"pledge-backend/db"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthAndEtcService struct{}

func NewEthAndEtc() *EthAndEtcService {
	return &EthAndEtcService{}
}

var ethClient *ethclient.Client

func init() {
	// 初始化以太坊客户端
	client, err := ethclient.Dial(config.Config.ETH.EthereumNodeURL)
	if err != nil {
		log.Fatal(err)
	}
	ethClient = client
}

// 获取区块数据
func (c *EthAndEtcService) GetBlockInfo(blockNumber *big.Int, tag string, full bool) (*types.Block, error) {
	var block *types.Block
	// var err error
	ctx := context.Background()

	//含有 head 、finalized、safe 其中某个字符串
	if tag != "" {
		cacheKey := fmt.Sprintf("block:%s", tag)
		//从redis中获取
		redisTokenInfoBytes, err := db.RedisGet(cacheKey)
		if err != nil {
			//redis 中不存在head、finalized、safe数据
			//从链上获取
			switch tag {
			case "head":
				block, err = ethClient.BlockByNumber(ctx, nil)
			case "finalized":
				block, err = ethClient.BlockByNumber(ctx, big.NewInt(-3))
			case "safe":
				block, err = ethClient.BlockByNumber(ctx, big.NewInt(-4))
			}
			if err := json.Unmarshal(redisTokenInfoBytes, &block); err != nil {
				return nil, err
			}
			if err != nil {
				return nil, err
			}
			// 保存到Redis和MySQL
			db.RedisSet(cacheKey, block, 10)
			models.SaveBlockToMySQL(block, full)
			return block, nil
		} else {
			return block, nil
		}
	} else {
		//区块号
		// 先尝试从数据库获取
		// var dbBlock models.Block
		block, err := models.GetBlockFromMySQL(blockNumber, full)
		if err == nil {
			return block, err
		}

		// 从链上获取
		block, err = ethClient.BlockByNumber(context.Background(), blockNumber)
		if err != nil {
			return nil, fmt.Errorf("获取区块失败: %v", err)
		}

		// 关键修复：检查 block 是否为 nil
		if block == nil {
			return nil, fmt.Errorf("区块 %s 不存在", blockNumber.String())
		}

		// 保存到数据库
		if err := models.SaveBlockToMySQL(block, full); err != nil {
			log.Printf("Failed to save block to DB: %v", err)
		}
		return block, nil
	}
}

//  获取交易数据
func (c *EthAndEtcService) GetTx(txHash string) (*models.Transaction, error) {
	// 1. 尝试从数据库获取
	dbTx, err := models.GetTxFromMySQL(txHash)
	if err != nil {
		return nil, err
	}
	if dbTx != nil {
		return dbTx, nil
	}

	// 如果数据库中没有，从区块链获取
	tx, _, err := ethClient.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, err
	}

	receipt, err := ethClient.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, err
	}

	// 准备交易数据
	transaction := &models.Transaction{
		TxHash:         txHash,
		BlockNumber:    receipt.BlockNumber.Int64(),
		ToAddress:      tx.To().Hex(),
		Value:          tx.Value().Int64(),
		GasPrice:      tx.GasPrice().Int64(),
		Data:     		common.Bytes2Hex(tx.Data()),
	}
	// 保存到数据库
	err = models.SaveTxFromMySQL(transaction)
	if err != nil {
		return nil, err
	}
	return transaction,err
}

//获取存根数据
func (c *EthAndEtcService) GetTxReceipt(txHash string) (*models.EthTxReceipt, error) {
	// 1. 尝试从数据库获取
	ethTxReceipt, err := models.GetByTxHash(txHash)
	if err != nil {
		return nil, err
	}
	if ethTxReceipt != nil {
		return ethTxReceipt, nil
	}

	// 如果数据库中没有，从区块链获取
	receipt, err := ethClient.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, err
	}

	// 获取交易详情以获取from/to地址
	tx, _, err := ethClient.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, err
	}

	// 准备收据数据
	ethReceipt := &models.EthTxReceipt{
		TxHash:          txHash,
		BlockHash:       receipt.BlockHash.Hex(),
		BlockNumber:     receipt.BlockNumber.Uint64(),
		ContractAddress: receipt.ContractAddress.Hex(),
		GasUsed:        receipt.GasUsed,
		Status:         receipt.Status,
		From:           "", // 需要通过其他方式获取
		To:             tx.To().Hex(),
		LogsCount:      len(receipt.Logs),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 获取发送方地址
	from, err := ethClient.TransactionSender(context.Background(), tx, receipt.BlockHash, receipt.TransactionIndex)
	if err == nil {
		ethReceipt.From = from.Hex()
	}

	// 将完整收据数据转为JSON存储
	receiptData, err := json.Marshal(receipt)
	if err == nil {
		ethReceipt.ReceiptData = string(receiptData)
	}

	// 保存到数据库
	err = models.SaveTxReceipt(ethReceipt)
	if err != nil {
		return nil, err
	}

	return ethReceipt, nil
}