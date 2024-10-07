package biz

import (
	"errors"
	"time"

	"github.com/ohko/chatroom/config"
)

func GroupList() (list []config.TableGroup, err error) {
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	err = tx.Find(&list).Error
	return
}

func GroupListByID(ids []int) (list []config.TableGroup, err error) {
	tx := config.DB.Begin()
	defer tx.Rollback()

	err = tx.Where("group_id IN ?", ids).Find(&list).Error
	return
}

func GroupDetail(GroupID int) (info config.TableGroup, err error) {
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	err = tx.First(&info, GroupID).Error
	return
}

func GroupCreate(info *config.TableGroup, userIds []int) error {
	if info.GroupName == "" || info.CreateUserID == 0 || info.OwnerID == 0 {
		return errors.New("GroupName/CreateUserID/OwnerID is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	info.GroupID = 0
	info.CreateTime = time.Now()

	tx := config.DB.Begin()
	defer tx.Rollback()

	if err := tx.Create(info).Error; err != nil {
		return err
	}

	for _, id := range userIds {
		if err := tx.Create(&config.TableUserGroup{UserID: id, GroupID: info.GroupID, JoinTime: time.Now()}).Error; err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func GroupUpdate(info *config.TableGroup) error {
	if info.GroupName == "" {
		return errors.New("GroupName is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	info.UpdateTime = time.Now()

	tx := config.DB.Begin()
	defer tx.Rollback()

	if err := tx.Save(info).Error; err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func GroupDelete(id int) error {
	if id == 0 {
		return errors.New("GroupID is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	group := config.TableGroup{GroupID: id}
	if err := tx.Delete(&group).Error; err != nil {
		return err
	}

	userGroup := config.TableUserGroup{GroupID: id}
	if err := tx.Where(&userGroup).Delete(&userGroup).Error; err != nil {
		return err
	}

	message := config.TableMessage{GroupID: id}
	if err := tx.Where(&message).Delete(&message).Error; err != nil {
		return err
	}

	tx.Commit()
	return nil
}
