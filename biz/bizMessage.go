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
	if err = tx.Model(&config.TableUser{}).
		Select(`user_id, account, real_name, avatar`).
		Where("user_id!=?", userID).
		Find(&users).Error; err != nil {
		return
	}

	{
		var msgs []config.TableMessage
		subQuery := tx.Model(&config.TableMessage{}).Select("*, ROW_NUMBER() OVER (PARTITION BY from_user_id, to_user_id ORDER BY message_id DESC) as rn_user").Where("group_id=0")
		if err = tx.Preload("FromUser").Table("(?) as a", subQuery).Where(`rn_user=1`).Find(&msgs).Error; err != nil {
			return
		}
		msgsIndex := map[any]config.TableMessage{}
		for _, m := range msgs {
			if m.FromUserID <= m.ToUserID {
				msgsIndex[fmt.Sprintf("0::%v::%v", m.FromUserID, m.ToUserID)] = m
			} else {
				msgsIndex[fmt.Sprintf("0::%v::%v", m.ToUserID, m.FromUserID)] = m
			}
		}
		for i, u := range users {
			key := ""
			if userID <= u.UserID {
				key = fmt.Sprintf("0::%v::%v", userID, u.UserID)
			} else {
				key = fmt.Sprintf("0::%v::%v", u.UserID, userID)
			}
			if m, ok := msgsIndex[key]; ok {
				users[i].LastMessage = &m
			}
		}
	}

	var groups []config.Contact
	if err = tx.Model(&config.TableUserGroup{}).
		Select(`g.group_id AS group_id,group_name, avatar, user_id`).
		Joins(`LEFT JOIN table_group g ON g.group_id=table_user_group.group_id`).
		Where(`user_id=?`, userID).
		Find(&groups).Error; err != nil {
		return
	}

	{
		var msgs []config.TableMessage
		subQuery := tx.Model(&config.TableMessage{}).Select("*, ROW_NUMBER() OVER (PARTITION BY group_id ORDER BY message_id DESC) as rn_group").Where("group_id!=0")
		if err = tx.Preload("FromUser").Table("(?) as a", subQuery).Where(`rn_group=1`).Find(&msgs).Error; err != nil {
			return
		}
		msgsIndex := map[int]config.TableMessage{}
		for _, m := range msgs {
			msgsIndex[m.GroupID] = m
		}
		for i, g := range groups {
			if m, ok := msgsIndex[g.GroupID]; ok {
				groups[i].LastMessage = &m
			}
			groups[i].UserID = 0
		}
	}

	list = append(users, groups...)
	return
}

func ContactsAndLastMessageByAccount(account string) (list []config.Contact, err error) {
	info := config.TableUser{Account: account}

	if err = func() error {
		config.DBLock.Lock()
		defer config.DBLock.Unlock()

		tx := config.DB.Begin()
		defer tx.Rollback()

		return tx.Where(&info).First(&info).Error
	}(); err != nil {
		return
	}

	return ContactsAndLastMessage(info.UserID)
}

func MessageList(groupID, FromUserID, ToUserID, offset, limit int) (list []config.TableMessage, err error) {
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	if groupID != 0 {
		err = tx.Preload("FromUser").Where("group_id=?", groupID).Order("message_id DESC").Offset(offset).Limit(limit).Find(&list).Error
	} else {
		err = tx.Preload("FromUser").Where("group_id=0 AND ((from_user_id=? AND to_user_id=?) OR (from_user_id=? AND to_user_id=?))", FromUserID, ToUserID, ToUserID, FromUserID).Order("message_id DESC").Offset(offset).Limit(limit).Find(&list).Error
	}

	ids := []int{}
	for _, l := range list {
		ids = append(ids, l.MessageID)
	}

	if len(ids) != 0 {
		err = tx.Model(&config.TableMessage{}).Where("is_read=0 AND message_id IN ?", ids).UpdateColumn("is_read", 1).Error
	}

	if err != nil {
		return
	}

	reversed := make([]config.TableMessage, len(list))
	for i := range list {
		reversed[i] = list[len(list)-1-i]
	}
	list = reversed

	tx.Commit()
	return
}

func MessageSend(info *config.TableMessage, wsToUserFunc func(userID int, info *config.TableMessage)) error {
	if err := messageSend(info); err != nil {
		return err
	}
	if wsToUserFunc != nil {
		if info.GroupID != 0 {
			if userGroups, err := UserGroupListByGroupID(info.GroupID); err == nil {
				for _, ug := range userGroups {
					info.ToUserID = ug.UserID
					wsToUserFunc(ug.UserID, info)
				}
			}
		} else {
			wsToUserFunc(info.ToUserID, info)
		}
	}
	return nil
}

func messageSend(info *config.TableMessage) error {
	if info.FromUserID == 0 || info.Type == "" || info.Content == "" {
		return errors.New("from_user_id/type/content is empty")
	}
	if info.ToUserID == 0 && info.GroupID == 0 {
		return errors.New("group_id or to_user_id both empty")
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
