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

	{ // last user message
		var msgs []config.TableMessage
		subQuery := tx.Model(&config.TableMessage{}).Select("*, ROW_NUMBER() OVER (PARTITION BY from_user_id, to_user_id ORDER BY message_id DESC) as rn_user").Where("group_id=0 AND (to_user_id=? OR from_user_id=?)", userID, userID)
		if err = tx.Preload("FromUser").Table("(?) as a", subQuery).Where(`rn_user=1`).Find(&msgs).Error; err != nil {
			return
		}
		msgsIndex := map[any]config.TableMessage{}
		for _, m := range msgs {
			var key string
			if m.FromUserID <= m.ToUserID {
				key = fmt.Sprintf("0::%v::%v", m.FromUserID, m.ToUserID)
			} else {
				key = fmt.Sprintf("0::%v::%v", m.ToUserID, m.FromUserID)
			}
			if v, ok := msgsIndex[key]; !ok || v.MessageID < m.MessageID {
				msgsIndex[key] = m
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
				if m.IsRead == 0 && m.ToUserID == userID {
					users[i].UnRead = 1
				}
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

	{ // last group message
		var msgs []config.TableMessage
		subQuery := tx.Model(&config.TableMessage{}).Select("*, ROW_NUMBER() OVER (PARTITION BY group_id ORDER BY message_id DESC) as rn_group").Where("group_id!=0 AND to_user_id=?", userID)
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
				if m.IsRead == 0 && m.ToUserID == userID {
					groups[i].UnRead = 1
				}
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
		err = tx.Preload("FromUser").Where("group_id=? AND to_user_id=?", groupID, ToUserID).Order("message_id DESC").Offset(offset).Limit(limit).Find(&list).Error
	} else {
		err = tx.Preload("FromUser").Where("group_id=0 AND ((from_user_id=? AND to_user_id=?) OR (from_user_id=? AND to_user_id=?))", FromUserID, ToUserID, ToUserID, FromUserID).Order("message_id DESC").Offset(offset).Limit(limit).Find(&list).Error
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
	go MessageRead(FromUserID, ToUserID, groupID)
	return
}

func MessageLastList(userIDs []int) (list []config.TableMessage, err error) {
	if len(userIDs) == 0 {
		return nil, errors.New("UserID is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	// last user message
	var msgs1 []config.TableMessage
	subQuery1 := tx.Model(&config.TableMessage{}).Select("*, ROW_NUMBER() OVER (PARTITION BY from_user_id, to_user_id ORDER BY message_id DESC) as rn_user").Where("group_id=0 AND to_user_id IN ?", userIDs)
	if err = tx.Preload("FromUser").Table("(?) as a", subQuery1).Where(`rn_user=1`).Find(&msgs1).Error; err != nil {
		return
	}

	// last group message
	var msgs2 []config.TableMessage
	subQuery2 := tx.Model(&config.TableMessage{}).Select("*, ROW_NUMBER() OVER (PARTITION BY group_id ORDER BY message_id DESC) as rn_group").Where("group_id!=0 AND to_user_id IN ?", userIDs)
	if err = tx.Preload("FromUser").Table("(?) as a", subQuery2).Where(`rn_group=1`).Find(&msgs2).Error; err != nil {
		return
	}

	list = append(msgs1, msgs2...)
	return
}

func MessageSend(info *config.TableMessage) error {
	if info.FromUserID == 0 || info.Type == "" || info.Content == "" {
		return errors.New("FromUserID/Type/Content is empty")
	}
	if info.ToUserID == 0 && info.GroupID == 0 {
		return errors.New("GroupID or ToUserID do not both empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	info.CreateTime = time.Now()

	tx := config.DB.Begin()
	defer tx.Rollback()

	if info.GroupID != 0 {
		var list []config.TableUserGroup
		if err := tx.Where(&config.TableUserGroup{GroupID: info.GroupID}).Find(&list).Error; err != nil {
			return err
		}
		for _, l := range list {
			info.MessageID = 0
			info.ToUserID = l.UserID
			if err := tx.Create(&info).Error; err != nil {
				return err
			}
			if err := tx.First(&info.FromUser, info.FromUserID).Error; err != nil {
				return err
			}
			info.SenderExtData = ""
			WsSendMessageByUserID(info.ToUserID, info)
		}
	} else {
		info.MessageID = 0
		if err := tx.Create(&info).Error; err != nil {
			return err
		}
		if err := tx.First(&info.FromUser, info.FromUserID).Error; err != nil {
			return err
		}
		info.SenderExtData = ""
		WsSendMessageByUserID(info.ToUserID, info)
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

func MessageRead(fromUserID, toUserID, groupID int) error {
	if toUserID == 0 {
		return errors.New("ToUserID is empty")
	}
	if fromUserID == 0 && groupID == 0 {
		return errors.New("FromUserID or GroupID do not both empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	if groupID != 0 {
		if err := tx.Model(&config.TableMessage{}).Where("is_read=0 AND group_id=? AND to_user_id=?", groupID, toUserID).Update("is_read", 1).Error; err != nil {
			return err
		}
	} else {
		if err := tx.Model(&config.TableMessage{}).Where("is_read=0 AND from_user_id=? AND to_user_id=?", fromUserID, toUserID).Update("is_read", 1).Error; err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}
