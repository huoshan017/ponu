package main

import (
	"sync"
)

type Account struct {
	account string
}

type AccountMgr struct {
	accounts map[string]*Account
	locker   sync.RWMutex
}

var account_mgr AccountMgr

func (this *AccountMgr) Init() {
	this.accounts = make(map[string]*Account)
}

func (this *AccountMgr) Add(account *Account) bool {
	this.locker.Lock()
	defer this.locker.Unlock()
	_, o := this.accounts[account.account]
	if o {
		return false
	}
	this.accounts[account.account] = account
	return true
}

func (this *AccountMgr) Remove(account string) bool {
	this.locker.Lock()
	defer this.locker.Unlock()
	_, o := this.accounts[account]
	if !o {
		return false
	}
	delete(this.accounts, account)
	return true
}

func (this *AccountMgr) Verify(account string, password string) bool {
	return true
}

func (this *AccountMgr) Num() int {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return len(this.accounts)
}
