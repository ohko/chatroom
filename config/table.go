package config

import "time"

type TableUser struct {
	UserID        int       `gorm:"user_id;primary_key"`
	Account       string    `gorm:"account;index:check;unique"`
	Password      string    `gorm:"password;index:check" json:"-"`
	RealName      string    `gorm:"real_name"`
	Avatar        string    `gorm:"avatar"`
	CreateTime    time.Time `gorm:"create_time"`
	UpdateTime    time.Time `gorm:"update_time"`
	LastLoginTime time.Time `gorm:"last_login_time"`
	LastLoginIP   string    `gorm:"last_login_ip"`
	RegisterIP    string    `gorm:"register_ip"`
}

type TableGroup struct {
	GroupID      int       `gorm:"group_id;primary_key"`
	GroupName    string    `gorm:"group_name"`
	Avatar       string    `gorm:"avatar"`
	CreateUserID int       `gorm:"create_user_id;index"`
	CreateTime   time.Time `gorm:"create_time"`
	UpdateTime   time.Time `gorm:"update_time"`
	OwnerID      int       `gorm:"owner_id;index"`
}

type TableUserGroup struct {
	UserGroupID int       `gorm:"user_group_id;primary_key"`
	UserID      int       `gorm:"user_id;uniqueIndex:ugunique"`
	GroupID     int       `gorm:"group_id;uniqueIndex:ugunique"`
	JoinTime    time.Time `gorm:"join_time"`
	// Unread      int       `gorm:"unread;comment:unread message number"`

	User  TableUser  `gorm:"references:UserID"`
	Group TableGroup `gorm:"references:GroupID"`
}

type TableMessage struct {
	MessageID  int       `gorm:"message_id;primary_key"`
	FromUserID int       `gorm:"from_user_id;index"`
	ToUserID   int       `gorm:"to_user_id;index"`
	GroupID    int       `gorm:"group_id;comment:0=normal/x=group"`
	Type       string    `gorm:"'type';comment:text/image"`
	Content    string    `gorm:"content"`
	IsRead     int       `gorm:"is_read;comment:0=normal/1=read"`
	IsUndo     int       `gorm:"is_undo;comment:0=normal/1=undo"`
	CreateTime time.Time `gorm:"create_time"`

	FromUser TableUser `gorm:"foreignKey:FromUserID"`
}
