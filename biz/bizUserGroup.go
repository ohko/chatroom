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

	err = tx.Where(&config.TableUserGroup{UserID: userID}).Find(&list).Error
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

	err = tx.Where(&config.TableUserGroup{GroupID: groupID}).Find(&list).Error
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

	for _, id := range userIDs {
		info := config.TableUserGroup{UserID: id, GroupID: groupID, JoinTime: time.Now()}
		if err := tx.Save(&info).Error; err != nil {
			return err
		}
	}

	tx.Commit()
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
	return nil
}
