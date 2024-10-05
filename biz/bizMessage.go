package biz

import (
	"errors"
	"fmt"
	"time"

	"github.com/ohko/chatroom/config"
)

func MessageContacts(userID int) (list []map[string]any, err error) {
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	var users []map[string]any
	if err = tx.Raw(`SELECT user_id AS ID,account AS Name,avatar AS Avatar, 0 AS IsGroup FROM table_user`).Find(&users).Error; err != nil {
		return
	}

	var groups []map[string]any
	if err = tx.Raw(`SELECT g.group_id AS ID,group_name AS Name,avatar AS Avatar, user_id AS UserID, 1 AS IsGroup
FROM table_user_group ug
LEFT JOIN table_group g ON g.group_id=ug.group_id
WHERE user_id=?`, userID).Find(&groups).Error; err != nil {
		return
	}

	var msgs []map[string]any
	if err = tx.Raw(`SELECT MAX(message_id),* FROM table_message
WHERE to_user_id=?
GROUP BY from_user_id,to_user_id,group_id`, userID).Find(&msgs).Error; err != nil {
		return
	}
	msgsIndex := map[any]map[string]any{}
	for _, m := range msgs {
		if m["group_id"].(int64) == 0 {
			msgsIndex[fmt.Sprintf("0::%v", m["from_user_id"])] = m
		} else {
			msgsIndex[fmt.Sprintf("%v::%v", m["group_id"], m["to_user_id"])] = m
		}
	}
	for i, u := range users {
		key := fmt.Sprintf("0::%v", u["ID"])
		message_id, content, is_read, message_time := 0, "", 0, time.Unix(0, 0)
		if m, ok := msgsIndex[key]; ok {
			message_id, content, is_read, message_time = int(m["message_id"].(int64)), m["content"].(string), int(m["is_read"].(int64)), m["create_time"].(time.Time)
		}
		users[i]["LastMessageID"] = message_id
		users[i]["LastMessageContent"] = content
		users[i]["LastMessageTime"] = message_time
		users[i]["IsRead"] = is_read
	}
	for i, g := range groups {
		key := fmt.Sprintf("%v::%v", g["ID"], g["UserID"])
		message_id, content, is_read, message_time := 0, "", 0, time.Unix(0, 0)
		if m, ok := msgsIndex[key]; ok {
			message_id, content, is_read, message_time = int(m["message_id"].(int64)), m["content"].(string), int(m["is_read"].(int64)), m["create_time"].(time.Time)
		}
		groups[i]["LastMessageID"] = message_id
		groups[i]["LastMessageContent"] = content
		groups[i]["LastMessageTime"] = message_time
		groups[i]["IsRead"] = is_read
		delete(groups[i], "UserID")
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
		err = tx.Where("group_id=?", groupID).Offset(offset).Limit(limit).Find(&list).Error
	} else {
		err = tx.Where("group_id=0 AND ((from_user_id=? AND to_user_id=?) OR (from_user_id=? AND to_user_id=?))", FromUserID, ToUserID, ToUserID, FromUserID).Offset(offset).Limit(limit).Find(&list).Error
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

func MessageSend(info *config.TableMessage) error {
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
