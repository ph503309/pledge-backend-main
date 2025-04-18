package controllers

import (
	"net/http"
	"pledge-backend/api/models/ws"
	"pledge-backend/log"
	"pledge-backend/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type PriceController struct {
}

func (c *PriceController) NewPrice(ctx *gin.Context) {

	defer func() {
		recoverRes := recover()
		if recoverRes != nil {
			log.Logger.Sugar().Error("new price recover ", recoverRes)
		}
	}()

	conn, err := (&websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 5 * time.Second,
		CheckOrigin: func(r *http.Request) bool { //Cross domain
			return true
		},
	}).Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Logger.Sugar().Error("websocket request err:", err)
		return
	}

	randomId := ""
	remoteIP := ctx.RemoteIP()
	if remoteIP != "" {
		randomId = strings.Replace(remoteIP, ".", "_", -1) + "_" + utils.GetRandomString(23)
	} else {
		randomId = utils.GetRandomString(32)
	}

	server := &ws.Server{
		Id:       randomId,
		Socket:   conn,
		Send:     make(chan []byte, 800),
		LastTime: time.Now().Unix(),
	}

	go server.ReadAndWrite()
}
