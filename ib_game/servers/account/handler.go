package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
)

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

	var iret int
	iret, err = w.Write(data)
	if nil != err {
		//_send_error(w, 0, -1)
		log.Printf("client_msg_handler write data 2 failed err[%s] ret %d\n", err.Error(), iret)
		return
	}
}
