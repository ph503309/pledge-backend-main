package main

import (
	"pledge-backend/db"
	"pledge-backend/schedule/models"
	"pledge-backend/schedule/tasks"
	"time"
)

func main() {

	// init mysql
	db.InitMysql()

	// init redis
	db.InitRedis()

	// create table
	models.InitTable()

	// pool task
	tasks.Task()
	// 启动定时任务
	blockTasks := []*tasks.BlockPollingTask{
		tasks.NewBlockPollingTask("head", 5*time.Second),
		tasks.NewBlockPollingTask("finalized", 5*time.Second),
		tasks.NewBlockPollingTask("safe", 5*time.Second),
	}
	for _, task := range blockTasks {
		task.Start()
	}

}

/*
 If you change the version, you need to modify the following files'
 config/init.go
*/
