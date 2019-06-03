package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/huoshan017/mysql-go/proxy/client"
	phttp "github.com/huoshan017/ponu/http"
	"github.com/huoshan017/ponu/ib_game/servers/account/account_db"
)

type Config struct {
	Id                int32
	Name              string
	ListenAddr        string
	DBProxyServerAddr string
	DBHostId          int32
	DBHostAlias       string
	DBName            string
}

func (this *Config) Init(config_path string) bool {
	data, err := ioutil.ReadFile(config_path)
	if err != nil {
		log.Printf("read config file err: %v\n", config_path, err.Error())
		return false
	}
	err = json.Unmarshal(data, this)
	if err != nil {
		log.Printf("json unmarshal err: %v\n", err.Error())
		return false
	}
	return true
}

func main() {
	if len(os.Args) < 2 {
		log.Printf("args not enough, must specify a config file for db define\n")
		return
	}

	arg_config_file := flag.String("c", "", "config file path")
	flag.Parse()

	var config_path string
	if nil != arg_config_file {
		config_path = *arg_config_file
		log.Printf("config file path %v\n", config_path)
	} else {
		log.Printf("not specified config file arg\n")
		return
	}

	var config Config
	if !config.Init(config_path) {
		return
	}

	var db mysql_proxy.DB
	if !db.Connect(config.DBProxyServerAddr, config.DBHostId, config.DBHostAlias, config.DBName) {
		return
	}

	var table_proxys account_db.TableProxysManager
	table_proxys.Init(&db)

	account_table_proxy := table_proxys.Get_T_Account_Table_Proxy()
	accounts := account_table_proxy.SelectAllPrimaryField()
	for i := 0; i < len(accounts); i++ {
		log.Printf("index: %v, account: %v\n", i, accounts[i])
	}

	var hs phttp.Service
	hs.HandleFunc("/account", request_handler)
	hs.GoRun(config.ListenAddr)

	for {
		time.Sleep(time.Millisecond * 100)
	}
}

func request_handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if nil != err {
		//_send_error(w, 0, -1)
		log.Printf("request_handler ReadAll err[%s]", err.Error())
		return
	}
}
