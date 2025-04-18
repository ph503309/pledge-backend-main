package models

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"pledge-backend/db"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

type Block struct {
	Number     int64     `gorm:"column:number"`
	ParentHash string    `gorm:"column:parent_hash"`
	Timestamp  time.Time `gorm:"column:timestamp"`
	TxHash     string    `gorm:"column:tx_hash"`
}

type Transaction struct {
	Nonce       uint64 `gorm:"column:nonce"`
	ToAddress   string `gorm:"column:to_address"`
	Value       int64  `gorm:"column:value"`
	GasLimit    int64  `gorm:"column:gas_limit"`
	GasPrice    int64  `gorm:"column:gas_price"`
	Data        string `gorm:"column:data"`
	TxHash      string `gorm:"uniqueIndex"`
	BlockNumber int64  `gorm:"column:block_number"` // 外键关联区块
}

type EthTxReceipt struct {
	TxHash          string    `gorm:"primaryKey;size:66" json:"tx_hash"`
	BlockHash       string    `gorm:"size:66" json:"block_hash"`
	BlockNumber     uint64    `json:"block_number"`
	ContractAddress string    `gorm:"size:42" json:"contract_address"`
	GasUsed         uint64    `json:"gas_used"`
	Status          uint64    `json:"status"`
	From            string    `gorm:"size:42" json:"from"`
	To              string    `gorm:"size:42" json:"to"`
	LogsCount       int       `json:"logs_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ReceiptData     string    `gorm:"type:text" json:"receipt_data"` // 存储完整的收据数据 JSON
}

func NewBlock() *Block {
	return &Block{}
}

func SaveBlockToMySQL(block *types.Block, full bool) error {
	// 转换区块头信息
	dbBlock := Block{
		Number:     block.Number().Int64(),
		ParentHash: block.ParentHash().Hex(),
		Timestamp:  time.Unix(int64(block.Time()), time.Hour.Milliseconds()),
		TxHash:     block.TxHash().Hex(),
	}
	// 1. 先保存区块
	if err := db.Mysql.Save(&dbBlock).Error; err != nil {
		return fmt.Errorf("保存区块失败: %v", err)
	}
	// 保存交易
	if full {
		var transactions []Transaction
		for _, ethTx := range block.Transactions() {
			// 构造交易模型
			txValue := Transaction{
				BlockNumber: dbBlock.Number,   // 使用已保存的区块号
				ToAddress:   ethTx.To().Hex(), // 注意可能是合约创建交易（nil）
				Value:       ethTx.Value().Int64(),
				Nonce:       ethTx.Nonce(),
				GasPrice:    ethTx.GasPrice().Int64(),
				GasLimit:    int64(ethTx.Gas()),
				Data:        hex.EncodeToString(ethTx.Data()),
			}

			// 处理合约创建交易（ToAddress 为空）
			if ethTx.To() == nil {
				txValue.ToAddress = "0x"
			}

			transactions = append(transactions, txValue)
		}

		// 批量插入交易
		if len(transactions) > 0 {
			if err := db.Mysql.CreateInBatches(transactions, 100).Error; err != nil {
				return fmt.Errorf("保存交易失败: %v", err)
			}
		}
	}
	return nil
}

func GetBlockFromMySQL(number *big.Int, full bool) (*types.Block, error) {
	var dbBlock Block
	err := db.Mysql.Table("block").Where("number = ?", number.Int64()).First(&dbBlock).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("区块不存在")
		} else {
			return nil, errors.New("record select err " + err.Error())
		}
	}
	// 2. 转换区块头
	ethHeader := &types.Header{
		Number:     big.NewInt(dbBlock.Number),
		ParentHash: common.HexToHash(dbBlock.ParentHash),
		Time:       uint64(dbBlock.Timestamp.Unix()), // 假设时间戳是 time.Time 类型
		TxHash:     common.HexToHash(dbBlock.TxHash),
	}
	ethBlock := types.NewBlockWithHeader(ethHeader)
	if full {
		var dbTransactions []Transaction
		// 查询关联交易
		if err := db.Mysql.Table("transaction").Where("block_number = ?", dbBlock.Number).Find(&dbTransactions).Error; err != nil {
			return nil, fmt.Errorf("查询交易失败: %v", err)
		}

		// 转换交易数据
		txs := make([]*types.Transaction, 0, len(dbTransactions))
		for _, dbTx := range dbTransactions {
			tx := types.NewTransaction(
				uint64(dbTx.Nonce),
				common.HexToAddress(dbTx.ToAddress),
				big.NewInt(dbTx.Value),
				uint64(dbTx.GasLimit),
				big.NewInt(dbTx.GasPrice),
				common.FromHex(dbTx.Data),
			)
			txs = append(txs, tx)
		}
		ethBlock = types.NewBlockWithHeader(ethHeader).WithBody(txs, nil) // 更新交易列表
	}
	return ethBlock, err
}

func GetTxFromMySQL(TxHash string) (*Transaction, error) {
	var transaction Transaction
	err := db.Mysql.Where("tx_hash = ?", TxHash).First(&transaction).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &transaction, nil
}

func SaveTxFromMySQL(transaction *Transaction) error{
	return db.Mysql.Create(transaction).Error
}

func GetByTxHash(txHash string) (*EthTxReceipt, error) {
	var receipt EthTxReceipt
	err := db.Mysql.Where("tx_hash = ?", txHash).First(&receipt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &receipt, nil
}

func SaveTxReceipt(receipt *EthTxReceipt) error {
	return db.Mysql.Create(receipt).Error
}
