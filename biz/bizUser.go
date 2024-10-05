package biz

import (
	"errors"
	"net/http"
	"time"

	"github.com/ohko/chatroom/common"
	"github.com/ohko/chatroom/config"
)

func UserList() (list []config.TableUser, err error) {
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	err = tx.Find(&list).Error
	return
}

func UserDetail(UserID int) (info config.TableUser, err error) {
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	err = tx.First(&info, UserID).Error
	return
}

func UserLogin(account, password string, r *http.Request) (info config.TableUser, err error) {
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	info.Account = account
	info.Password = common.Hash(password)
	if err = tx.Where(&info).First(&info).Error; err != nil {
		return
	}

	info.LastLoginTime = time.Now()
	if r != nil {
		info.LastLoginIP = common.GetRealIP(r)
	}

	if err = tx.Save(info).Error; err != nil {
		return
	}

	tx.Commit()
	return
}

func UserRegister(info *config.TableUser, r *http.Request) error {
	if info.Account == "" || info.Password == "" {
		return errors.New("account/password is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	info.UserID = 0
	info.CreateTime = time.Now()
	originPassword := info.Password
	info.Password = common.Hash(info.Password)
	if r != nil {
		info.RegisterIP = common.GetRealIP(r)
	}

	tx := config.DB.Begin()
	defer tx.Rollback()

	exist := config.TableUser{Account: info.Account}
	var count int64
	if err := tx.Model(&exist).Where(&exist).Count(&count).Error; err != nil {
		return err
	}

	if count != 0 {
		return errors.New("account already exists")
	}

	if err := tx.Create(info).Error; err != nil {
		return err
	}
	info.Password = originPassword

	tx.Commit()
	return nil
}

func UserUpdate(info *config.TableUser) error {
	if info.Account == "" {
		return errors.New("UserID/Account is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	info.UpdateTime = time.Now()
	if info.Password != "" {
		info.Password = common.Hash(info.Password)
	}

	tx := config.DB.Begin()
	defer tx.Rollback()

	if err := tx.Save(info).Error; err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func UserDelete(id int) error {
	if id == 0 {
		return errors.New("UserID is empty")
	}
	config.DBLock.Lock()
	defer config.DBLock.Unlock()

	tx := config.DB.Begin()
	defer tx.Rollback()

	user := &config.TableUser{UserID: id}
	if err := tx.Delete(&user).Error; err != nil {
		return err
	}

	userGroup := config.TableUserGroup{UserID: id}
	if err := tx.Where(&userGroup).Delete(&userGroup).Error; err != nil {
		return err
	}

	message1 := config.TableMessage{ToUserID: id}
	if err := tx.Where(&message1).Delete(&message1).Error; err != nil {
		return err
	}
	message2 := config.TableMessage{FromUserID: id}
	if err := tx.Where(&message2).Delete(&message2).Error; err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func UserDeleteByAccount(account string) error {
	if account == "" {
		return errors.New("Account is empty")
	}
	info := config.TableUser{Account: account}

	if err := func() error {
		config.DBLock.Lock()
		defer config.DBLock.Unlock()

		tx := config.DB.Begin()
		defer tx.Rollback()

		return tx.Where(&info).First(&info).Error
	}(); err != nil {
		return err
	}

	return UserDelete(info.UserID)
}
