package controllers

import (
	"math/big"
	"net/http"
	"pledge-backend/api/services"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type EthAndEtcController struct {
}

// 获取区块数据, full是query参数，默认为false，是true时获取区块并包含交易数据
// block_num除了是数字以外，还可以单独处理 head、finalized、safe三个字符串
// 当查询head、finalized、safe区块时，尝试redis中获取，获取不到才会从链上获取
// 尝试把区块数据保存到mysql，
// 当full = true时，还需要把交易数据保存到mysql
// 如果mysql里面没有 才会去链上查询
func (c *EthAndEtcController) GetBlock(ctx *gin.Context) {
	blockParam := ctx.Param("block_num")
	full, _ := strconv.ParseBool(ctx.DefaultQuery("full", "false"))

	var blockNumber *big.Int
	var tag string
	if blockParam == "head" || blockParam == "finalized" || blockParam == "safe" {
		tag = blockParam
	} else {
		num, err := strconv.ParseInt(blockParam, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid block number"})
			return
		}
		blockNumber = big.NewInt(num)
	}

	block, err := services.NewEthAndEtc().GetBlockInfo(blockNumber, tag, full)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, block)
}

// 获取交易数据
// 尝试把交易数据保存到mysql
// 如果mysql中没有才会去链上查询
func (c *EthAndEtcController) GetTx(ctx *gin.Context) {
	txHash := ctx.Param("tx_hash")
	if !common.IsHexAddress(txHash) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction hash"})
		return
	}

	transaction, err := services.NewEthAndEtc().GetTx(txHash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, transaction)

}


// 尝试把存根数据保存到mysql。
// 如果mysql中没有，才会使用ethclient根据交易哈希去链上查询Receipt数据。
func (c *EthAndEtcController) GetTxReceipt(ctx *gin.Context) {
	txHash := ctx.Param("tx_hash")
	if !common.IsHexAddress(txHash) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction hash"})
		return
	}

	block, err := services.NewEthAndEtc().GetTxReceipt(txHash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, block)
}
