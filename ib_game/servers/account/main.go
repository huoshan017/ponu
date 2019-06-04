package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	phttp "github.com/huoshan017/ponu/http"
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

type Server struct {
	db_proxy     DBProxy
	http_service phttp.Service
	config       *Config
}

var server Server

func (this *Server) Init(config *Config) bool {
	if !this.db_proxy.Connect(config.DBProxyServerAddr, config.DBHostId, config.DBHostAlias, config.DBName) {
		return false
	}
	this.http_service.HandleFunc("/account", request_handler)
	this.config = config
	return true
}

func (this *Server) Run() {
	this.db_proxy.GoRun()
	this.http_service.GoRun(this.config.ListenAddr)
	for {
		time.Sleep(time.Millisecond * 100)
	}
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

	if !server.Init(&config) {
		return
	}

	server.Run()
}
