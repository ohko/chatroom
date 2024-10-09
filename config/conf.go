package config

import (
	"sync"

	"gorm.io/gorm"
)

var (
	DB        *gorm.DB
	DBLock    myLock
	AESKey    = "1234567887654321"
	TokenName = "token"
	DBType    string
)

type myLock struct {
	lock sync.RWMutex
}

func (o *myLock) Lock() {
	if DBType != "sqlite" {
		return
	}
	o.lock.Lock()
}

func (o *myLock) Unlock() {
	if DBType != "sqlite" {
		return
	}
	o.lock.Unlock()
}
