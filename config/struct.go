package config

import (
	"context"
)

type Context struct {
	Ctx      context.Context
	FlowID   string
	PostData string
}

type JSONData struct {
	No   int `json:"no"`
	Data any `json:"data"`
}

type ResultData struct {
	Error    error
	Type     string
	Template []string
	Data     any
}

type Contact struct {
	UserID      int
	GroupID     int
	Account     string
	RealName    string
	GroupName   string
	Avatar      string
	LastMessage *TableMessage `gorm:"-"`
}
