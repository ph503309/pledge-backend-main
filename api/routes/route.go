package routes

import (
	"pledge-backend/api/controllers"
	"pledge-backend/api/middlewares"
	"pledge-backend/config"
	"time"

	"github.com/gin-gonic/gin"
)

func InitRoute(e *gin.Engine) *gin.Engine {

	// version group
	v2Group := e.Group("/api/v" + config.Config.Env.Version)

	// pledge-defi backend
	poolController := controllers.PoolController{}
	v2Group.GET("/poolBaseInfo", poolController.PoolBaseInfo)                                   //pool base information
	v2Group.GET("/poolDataInfo", poolController.PoolDataInfo)                                   //pool data information
	v2Group.GET("/token", poolController.TokenList)                                             //pool token information
	v2Group.POST("/pool/debtTokenList", middlewares.CheckToken(), poolController.DebtTokenList) //pool debtTokenList
	v2Group.POST("/pool/search", middlewares.CheckToken(), poolController.Search)               //pool search

	// plgr-usdt price
	priceController := controllers.PriceController{}
	v2Group.GET("/price", priceController.NewPrice) //new price on ku-coin-exchange

	// pledge-defi admin backend
	multiSignPoolController := controllers.MultiSignPoolController{}
	v2Group.POST("/pool/setMultiSign", middlewares.CheckToken(), multiSignPoolController.SetMultiSign) //multi-sign set
	v2Group.POST("/pool/getMultiSign", middlewares.CheckToken(), multiSignPoolController.GetMultiSign) //multi-sign get

	userController := controllers.UserController{}
	v2Group.POST("/user/login", userController.Login)                             // login
	v2Group.POST("/user/logout", middlewares.CheckToken(), userController.Logout) // logout

	// 注册路由
	ethAndEtcController := controllers.EthAndEtcController{}
	v2Group.GET("/eth/block/:block_num",middlewares.RateLimiter(10,time.Minute), ethAndEtcController.GetBlock)
	v2Group.GET("/eth/tx/:tx_hash", middlewares.RateLimiter(10,time.Minute),ethAndEtcController.GetTx)
	v2Group.GET("/etc/tx_receipt/:tx_hash",middlewares.RateLimiter(10,time.Minute), ethAndEtcController.GetTxReceipt)
	return e
}
