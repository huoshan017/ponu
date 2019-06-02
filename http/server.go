package http

import (
	//"net"
	"net/http"
	"sync"
)

type Service struct {
	mux        *http.ServeMux
	err        error
	err_locker sync.RWMutex
}

func (this *Service) Init() {
	this.mux = http.NewServeMux()
	this.err = nil
}

func (this *Service) HandleFunc(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	if this.mux == nil {
		this.mux = http.NewServeMux()
	}
	this.mux.HandleFunc(pattern, handler)
}

func (this *Service) Handle(pattern string, handler http.Handler) {
	if this.mux == nil {
		this.mux = http.NewServeMux()
	}
	this.mux.Handle(pattern, handler)
}

func (this *Service) Start(addr string) error {
	return http.ListenAndServe(addr, this.mux)
}

func (this *Service) RunLoop(addr string) {
	go func() {
		err := http.ListenAndServe(addr, this.mux)
		this.err_locker.Lock()
		this.err = err
		this.err_locker.Unlock()
	}()
}

func (this *Service) GetErr() error {
	this.err_locker.RLock()
	defer this.err_locker.RUnlock()
	return this.err
}
