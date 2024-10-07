package biz

import (
	"errors"
	"fmt"
	"time"

	"github.com/ohko/chatroom/config"
)

func ContactsAndLastMessage(userID int) (list []config.Contact, err error) {
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	var users []config.Contact
	if err = tx.Raw(`SELECT user_id, account, real_name, avatar FROM table_user`).Find(&users).Error; err != nil {
		return
	}

	var groups []config.Contact
	if err = tx.Raw(`SELECT g.group_id AS group_id,group_name AS account, avatar, user_id
FROM table_user_group ug
LEFT JOIN table_group g ON g.group_id=ug.group_id
WHERE user_id=?`, userID).Find(&groups).Error; err != nil {
		return
	}

	var msgs []config.TableMessage
	if err = tx.Preload("FromUser").Select(`MAX(message_id),*`).Where(`to_user_id=?`, userID).Group("message_id").Group("from_user_id").Group("to_user_id").Group("group_id").Find(&msgs).Error; err != nil {
		return
	}
	msgsIndex := map[any]config.TableMessage{}
	for _, m := range msgs {
		if m.GroupID == 0 {
			msgsIndex[fmt.Sprintf("0::%v", m.FromUserID)] = m
		} else {
			msgsIndex[fmt.Sprintf("%v::%v", m.GroupID, m.ToUserID)] = m
		}
	}
	for i, u := range users {
		key := fmt.Sprintf("0::%v", u.UserID)
		if m, ok := msgsIndex[key]; ok {
			users[i].LastMessage = &m
		}
	}
	for i, g := range groups {
		key := fmt.Sprintf("%v::%v", g.GroupID, g.UserID)
		if m, ok := msgsIndex[key]; ok {
			groups[i].LastMessage = &m
		}
	}

	list = append(users, groups...)
	return
}

func MessageList(groupID, FromUserID, ToUserID, offset, limit int) (list []config.TableMessage, err error) {
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	if groupID != 0 {
		err = tx.Preload("FromUser").Where("group_id=?", groupID).Offset(offset).Limit(limit).Find(&list).Error
	} else {
		err = tx.Preload("FromUser").Where("group_id=0 AND ((from_user_id=? AND to_user_id=?) OR (from_user_id=? AND to_user_id=?))", FromUserID, ToUserID, ToUserID, FromUserID).Offset(offset).Limit(limit).Find(&list).Error
	}

	ids := []int{}
	for _, l := range list {
		if l.ToUserID == ToUserID && l.IsRead == 0 {
			ids = append(ids, l.MessageID)
		}
	}

	if len(ids) != 0 {
		err = tx.Model(&config.TableMessage{}).Where("is_read=0 AND message_id IN ?", ids).UpdateColumn("is_read", 1).Error
	}

	if err != nil {
		return
	}

	tx.Commit()
	return
}

func MessageSend(info *config.TableMessage, wsToUserFunc func(userID int, info *config.TableMessage)) error {
	if info.GroupID != 0 {
		userGroups, err := UserGroupListByGroupID(info.GroupID)
		if err != nil {
			return err
		}
		for _, ug := range userGroups {
			info.ToUserID = ug.UserID
			if err := messageSend(info); err != nil {
				return err
			}
			if wsToUserFunc != nil {
				wsToUserFunc(ug.UserID, info)
			}
		}
	} else {
		if err := messageSend(info); err != nil {
			return err
		}
		if wsToUserFunc != nil {
			wsToUserFunc(info.ToUserID, info)
		}
	}
	return nil
}

func messageSend(info *config.TableMessage) error {
	if info.FromUserID == 0 || info.ToUserID == 0 || info.Type == "" || info.Content == "" {
		return errors.New("from_user_id/to_user_id/type/content is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	info.MessageID = 0
	info.CreateTime = time.Now()

	tx := config.DB.Begin()
	defer tx.Rollback()

	if err := tx.Create(&info).Error; err != nil {
		return err
	}

	if err := tx.First(&info.FromUser, info.FromUserID).Error; err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func MessageUndo(messageID int) error {
	if messageID == 0 {
		return errors.New("MessageID is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	info := config.TableMessage{MessageID: messageID}
	if err := tx.Model(&info).Where(&info).Update("is_undo", 1).Error; err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func MessageRead(messageID int) error {
	if messageID == 0 {
		return errors.New("MessageID is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	info := config.TableMessage{MessageID: messageID}
	if err := tx.Model(&info).Where(&info).Update("is_read", 1).Error; err != nil {
		return err
	}

	tx.Commit()
	return nil
}
