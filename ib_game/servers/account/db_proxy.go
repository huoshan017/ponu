package main

import (
	"github.com/huoshan017/mysql-go/proxy/client"
	"github.com/huoshan017/ponu/ib_game/servers/account/account_db"
)

type DBProxy struct {
	db           mysql_proxy.DB
	table_proxys account_db.TableProxysManager
}

func (this *DBProxy) Connect(proxy_addr string, db_host_id int32, db_host_alias, db_name string) bool {
	if !this.db.Connect(proxy_addr, db_host_id, db_host_alias, db_name) {
		return false
	}
	this.table_proxys.Init(&this.db)
	return true
}

func (this *DBProxy) GoRun() {
	this.db.GoRun()
}

func (this *DBProxy) Save() {
	this.db.Save()
}

func (this *DBProxy) End() {
	this.db.Close()
}

func (this *DBProxy) GetTableManager() *account_db.TableProxysManager {
	return &this.table_proxys
}
