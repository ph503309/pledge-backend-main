package tasks

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"pledge-backend/config"
	"pledge-backend/db"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockType string

const (
	BlockTypeHead      BlockType = "head"
	BlockTypeFinalized BlockType = "finalized"
	BlockTypeSafe      BlockType = "safe"
)


type BlockPollingTask struct {
	ctx      context.Context
	tag      string
	cancel   context.CancelFunc
	interval time.Duration
	blockCh    chan BlockType
	wg         sync.WaitGroup
	quitCh     chan struct{}
}

func NewBlockPollingTask(tag string, interval time.Duration) *BlockPollingTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &BlockPollingTask{
		ctx:      ctx,
		tag:      tag,
		cancel:   cancel,
		interval: interval,
		blockCh:    make(chan BlockType, 10),
		quitCh:     make(chan struct{}),
	}
}

func (t *BlockPollingTask) Start() {
	go t.eventLoop()
}

func (t *BlockPollingTask) Stop() {
	t.cancel()
	<-t.quitCh // 等待eventLoop退出
	close(t.blockCh)
}

func (t *BlockPollingTask) eventLoop() {
	defer t.wg.Done()
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	client, err := ethclient.Dial(config.Config.ETH.EthereumNodeURL)
	if err != nil {
		log.Fatal(err)
	}
	cacheKey := fmt.Sprintf("block:%s", t.tag)

	// var block *types.Block
	// var err error
	
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			var number *big.Int
			switch t.tag {
			case "head":
				number = nil
			case "finalized":
				number = big.NewInt(-3)
			case "safe":
				number = big.NewInt(-4)
			}
			block, err := client.BlockByNumber(context.Background(), number)
			if err != nil {
				log.Printf("Error fetching %s block: %v", t.tag, err)
				continue
			}
			// 保存到Redis
			if err := db.RedisSet(cacheKey, block, 10); err != nil {
				log.Printf("Error saving %s block to Redis: %v", t.tag, err)
			}
		case <-t.quitCh:
			return
		}
	}
}
