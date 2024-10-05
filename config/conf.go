package config

import (
	"sync"

	"gorm.io/gorm"
)

var (
	DB        *gorm.DB
	DBLock    sync.RWMutex
	AESKey    = "1234567887654321"
	TokenName = "token"
)
