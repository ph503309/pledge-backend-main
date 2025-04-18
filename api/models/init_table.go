package models

import "pledge-backend/db"

func InitTable() {
	db.Mysql.AutoMigrate(&MultiSign{})
	db.Mysql.AutoMigrate(&TokenInfo{})
	db.Mysql.AutoMigrate(&TokenList{})
	db.Mysql.AutoMigrate(&PoolData{})
	db.Mysql.AutoMigrate(&PoolBases{})
	db.Mysql.AutoMigrate(&Block{})
	db.Mysql.AutoMigrate(&Transaction{})
	db.Mysql.AutoMigrate(&EthTxReceipt{})
}
