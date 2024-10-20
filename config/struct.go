package config

import (
	"context"
	"time"
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

type WSMsg struct {
	Type          string // ping/pong/text/image
	Token         string `json:",omitempty"` // type=bind
	MessageID     int    `json:",omitempty"`
	FromUserID    int    `json:",omitempty"`
	ToUserID      int    `json:",omitempty"`
	GroupID       int    `json:",omitempty"`
	Content       string `json:",omitempty"`
	No            int    `json:",omitempty"`
	Data          any    `json:",omitempty"`
	CreateTime    time.Time
	SenderExtData string
	ExtData       string
}

type Contact struct {
	UserID      int
	GroupID     int
	Account     string
	RealName    string
	GroupName   string
	Avatar      string
	UnRead      int
	ExtData     string
	LastMessage *TableMessage `gorm:"-"`
}
