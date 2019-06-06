package main

import (
	"flag"
	"log"
	"os"
)

func get_args() (config_path string) {
	if len(os.Args) < 2 {
		log.Printf("args not enough, must specify a config file for db define\n")
		return
	}

	arg_config_file := flag.String("c", "", "config file path")
	flag.Parse()

	if nil != arg_config_file {
		config_path = *arg_config_file
		log.Printf("config file path %v\n", config_path)
	} else {
		log.Printf("not specified config file arg\n")
		return
	}
	return
}

func main() {
	config_path := get_args()
	if config_path == "" {
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
