package com

import (
	"log"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/ohko/chatroom/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

func Init(dbPath string) error {
	var err error

	if config.DB, err = NewDB(dbPath); err != nil {
		log.Fatal(err)
	}

	return nil
}

func NewDB(dbPath string) (*gorm.DB, error) {
	options := gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",                                // table name prefix, table for `User` would be `t_users`
			SingularTable: true,                              // use singular table name, table for `User` would be `user` with this option enabled
			NoLowerCase:   false,                             // skip the snake_casing of names
			NameReplacer:  strings.NewReplacer("CID", "Cid"), // use name replacer to change struct/field name before convert it to db name
		},
		NowFunc: func() time.Time {
			return time.Now()
		},
		SkipDefaultTransaction: true}

	var dsn gorm.Dialector
	if strings.HasPrefix(dbPath, "postgres://") {
		config.DBType = "postgres"
		dsn = postgres.Open(dbPath)
	} else {
		config.DBType = "sqlite"
		dsn = sqlite.Open(dbPath)
	}
	db, err := gorm.Open(dsn, &options)
	if err != nil {
		return nil, err
	}

	db.Use(dbresolver.Register(dbresolver.Config{ /* xxx */ }).
		SetConnMaxIdleTime(time.Hour).
		SetConnMaxLifetime(24 * time.Hour).
		SetMaxIdleConns(100).
		SetMaxOpenConns(200))

	if err := db.AutoMigrate(
		&config.TableUserGroup{},
		&config.TableGroup{},
		&config.TableUser{},
		&config.TableMessage{},
	); err != nil {
		log.Fatal(err)
	}

	// if runtime.GOOS == "darwin" {
	// 	return db.Debug(), nil
	// }
	return db, nil
}
