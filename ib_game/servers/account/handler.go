package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
)

func verify_handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			log.Printf("%v\n", reflect.TypeOf(err))
		}
	}()

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if nil != err {
		//_send_error(w, 0, -1)
		log.Printf("verify handler ReadAll err %v\n", err.Error())
		return
	}

	data, err = _verify(data)
	if err != nil {

	}

	var ret int
	ret, err = w.Write(data)
	if nil != err {
		//_send_error(w, 0, -1)
		log.Printf("verify handler Write err %v, ret %v\n", err.Error(), ret)
		return
	}
}

func _verify(data []byte) (ret_data []byte, err error) {
	return
}

func register_handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			log.Printf("%v\n", reflect.TypeOf(err))
		}
	}()
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("register handler ReadAll err %v\n", err.Error())
		return
	}

	var ret int
	ret, err = w.Write(data)
	if err != nil {
		log.Printf("register handler Write err %v, ret %v\n", err.Error(), ret)
		return
	}
}
