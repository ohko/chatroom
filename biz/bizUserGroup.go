package biz

import (
	"errors"
	"time"

	"github.com/ohko/chatroom/config"
)

func UserGroupList(userID int) (list []config.TableUserGroup, err error) {
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	err = tx.Preload("User").Preload("Group").Where(&config.TableUserGroup{UserID: userID}).Find(&list).Error
	return
}

func UserGroupListByGroupID(groupID int) (list []config.TableUserGroup, err error) {
	if groupID == 0 {
		return nil, errors.New("GroupID is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	err = tx.Preload("User").Preload("Group").Where(&config.TableUserGroup{GroupID: groupID}).Find(&list).Error
	return
}

func UserGroupjoin(userIDs []int, groupID int) error {
	if groupID == 0 {
		return errors.New("UserID/GroupID is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	var count int64
	for _, id := range userIDs {
		info := config.TableUserGroup{UserID: id, GroupID: groupID}
		if err := tx.Model(&info).Where(&info).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			info.JoinTime = time.Now()
			if err := tx.Create(&info).Error; err != nil {
				return err
			}
			tx.Commit()
		}
	}

	go WsNotifyUserGroupJoin(groupID, userIDs)
	return nil
}

func UserGroupRemove(userIDs []int, groupID int) error {
	if groupID == 0 {
		return errors.New("UserID/GroupID is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	for _, id := range userIDs {
		if err := tx.Delete(&config.TableUserGroup{}, &config.TableUserGroup{UserID: id, GroupID: groupID}).Error; err != nil {
			return err
		}
	}

	tx.Commit()
	go WsNotifyUserGroupRemove(groupID, userIDs)
	return nil
}
