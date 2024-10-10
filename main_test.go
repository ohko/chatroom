package main

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/ohko/chatroom/biz"
	"github.com/ohko/chatroom/common"
	"github.com/ohko/chatroom/common/com"
	"github.com/ohko/chatroom/config"
)

// go test -timeout 1h -run ^Test_user$ github.com/ohko/chatroom -v -count=1
func Test_user(t *testing.T) {
	testUser := config.TableUser{Account: "user001", Password: "pass001"}

	com.Init(*dbPath)

	biz.UserDeleteByAccount(testUser.Account)

	if err := biz.UserRegister(&testUser, nil); err != nil {
		t.Fatal(err)
	}
	t.Log("Register success:", testUser.UserID)

	login, err := biz.UserLogin(testUser.Account, testUser.Password, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Login success:", login.UserID)

	list, err := biz.UserList()
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range list {
		t.Log(l.UserID, l.Account)
	}

	if err := biz.UserDelete(login.UserID); err != nil {
		t.Fatal(err)
	}
}

// go test -timeout 1h -run ^Test_group$ github.com/ohko/chatroom -v -count=1
func Test_group(t *testing.T) {
	com.Init(*dbPath)

	testUser := config.TableUser{Account: "user001", Password: "pass001"}
	biz.UserUpdate(&testUser)
	defer biz.UserDelete(testUser.UserID)
	t.Log("user_id:", testUser.UserID)

	testGroup := config.TableGroup{GroupName: "group001", CreateUserID: testUser.UserID, OwnerID: testUser.UserID}

	if err := biz.GroupCreate(&testGroup, []int{}); err != nil {
		t.Fatal(err)
	}

	t.Log("Group:", testGroup.GroupID)

	list, err := biz.GroupList()
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range list {
		t.Log(l.GroupID, l.GroupName)
	}

	if err := biz.GroupDelete(testGroup.GroupID); err != nil {
		t.Fatal(err)
	}
}

// go test -timeout 1h -run ^Test_usergroup$ github.com/ohko/chatroom -v -count=1
func Test_usergroup(t *testing.T) {
	com.Init(*dbPath)

	testUser := config.TableUser{Account: "user001", Password: "pass001"}
	biz.UserUpdate(&testUser)
	defer biz.UserDelete(testUser.UserID)
	t.Log("user_id:", testUser.UserID)

	testGroup := config.TableGroup{GroupName: "group001", CreateUserID: testUser.UserID, OwnerID: testUser.UserID}
	biz.GroupCreate(&testGroup, []int{})
	defer biz.GroupDelete(testGroup.GroupID)

	if err := biz.UserGroupjoin([]int{testUser.UserID}, testGroup.GroupID); err != nil {
		t.Fatal(err)
	}

	list, err := biz.UserGroupList(testUser.UserID)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range list {
		t.Log(l.UserGroupID, l.UserID, l.GroupID, l.JoinTime)
	}

	if err := biz.UserGroupRemove([]int{testUser.UserID}, testGroup.GroupID); err != nil {
		t.Fatal(err)
	}
}

// go test -timeout 1h -run ^Test_message$ github.com/ohko/chatroom -v -count=1
func Test_message(t *testing.T) {
	com.Init(*dbPath)

	biz.UserDeleteByAccount("fromUser")
	biz.UserDeleteByAccount("toUser")
	fromUser := config.TableUser{Account: "fromUser", Password: "fromUser"}
	biz.UserUpdate(&fromUser)
	defer biz.UserDelete(fromUser.UserID)
	t.Log("from_user_id:", fromUser.UserID)
	toUser := config.TableUser{Account: "toUser", Password: "toUser"}
	biz.UserUpdate(&toUser)
	defer biz.UserDelete(toUser.UserID)
	t.Log("to_user_id:", toUser.UserID)

	testGroup := config.TableGroup{GroupName: "group001", CreateUserID: fromUser.UserID, OwnerID: fromUser.UserID}
	biz.GroupCreate(&testGroup, []int{})
	defer biz.GroupDelete(testGroup.GroupID)

	biz.UserGroupjoin([]int{fromUser.UserID, toUser.UserID}, testGroup.GroupID)
	defer biz.UserGroupRemove([]int{fromUser.UserID, toUser.UserID}, testGroup.GroupID)

	message := config.TableMessage{
		FromUserID: fromUser.UserID,
		ToUserID:   toUser.UserID,
		GroupID:    testGroup.GroupID,
		Type:       "text",
		Content:    "hello",
		CreateTime: time.Now(),
	}
	if err := biz.MessageSend(&message); err != nil {
		t.Fatal(err)
	}

	list, err := biz.MessageList(testGroup.GroupID, toUser.UserID, toUser.UserID, 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range list {
		t.Log(l.GroupID, l.FromUserID, l.ToUserID, l.Content)
	}

	if err := biz.MessageUndo(message.MessageID); err != nil {
		t.Fatal(err)
	}
	if err := biz.MessageRead(fromUser.UserID, toUser.UserID, testGroup.GroupID); err != nil {
		t.Fatal(err)
	}
}

// go test -timeout 1h -run ^Test_lastMessage$ github.com/ohko/chatroom -v -count=1
func Test_lastMessage(t *testing.T) {
	com.Init(*dbPath)

	biz.UserDeleteByAccount("fromUser")
	biz.UserDeleteByAccount("toUser")
	fromUser := config.TableUser{Account: "fromUser", Password: "fromUser"}
	biz.UserUpdate(&fromUser)
	defer biz.UserDelete(fromUser.UserID)
	t.Log("from_user_id:", fromUser.UserID)
	toUser := config.TableUser{Account: "toUser", Password: "toUser"}
	biz.UserUpdate(&toUser)
	defer biz.UserDelete(toUser.UserID)
	t.Log("to_user_id:", toUser.UserID)

	{ // simaple
		message := config.TableMessage{FromUserID: fromUser.UserID, ToUserID: toUser.UserID, GroupID: 0, Type: "text", Content: "f=>t", CreateTime: time.Now()}
		biz.MessageSend(&message)
		list, _ := biz.ContactsAndLastMessage(fromUser.UserID)
		for _, l := range list {
			if l.Account != toUser.Account {
				continue
			}
			if l.LastMessage.Content != message.Content {
				t.Fatal(l.LastMessage)
			}
		}
		message = config.TableMessage{FromUserID: toUser.UserID, ToUserID: fromUser.UserID, GroupID: 0, Type: "text", Content: "t=>f", CreateTime: time.Now()}
		biz.MessageSend(&message)
		list, _ = biz.ContactsAndLastMessage(fromUser.UserID)
		for _, l := range list {
			if l.Account != toUser.Account {
				continue
			}
			if l.LastMessage.Content != message.Content {
				t.Fatal(l.LastMessage)
			}
		}
	}

	testGroup := config.TableGroup{GroupName: "group001", CreateUserID: fromUser.UserID, OwnerID: fromUser.UserID}
	biz.GroupCreate(&testGroup, []int{})
	defer biz.GroupDelete(testGroup.GroupID)
	biz.UserGroupjoin([]int{fromUser.UserID, toUser.UserID}, testGroup.GroupID)
	defer biz.UserGroupRemove([]int{fromUser.UserID, toUser.UserID}, testGroup.GroupID)
	{ // group
		message := config.TableMessage{FromUserID: fromUser.UserID, GroupID: testGroup.GroupID, Type: "text", Content: "f=>t", CreateTime: time.Now()}
		biz.MessageSend(&message)
		list, _ := biz.ContactsAndLastMessage(fromUser.UserID)
		for _, l := range list {
			if l.GroupID != testGroup.GroupID {
				continue
			}
			t.Log(l.LastMessage.Content)
			if l.LastMessage.Content != message.Content {
				t.Fatal(l.LastMessage)
			}
		}
		message = config.TableMessage{FromUserID: toUser.UserID, GroupID: testGroup.GroupID, Type: "text", Content: "t=>f", CreateTime: time.Now()}
		biz.MessageSend(&message)
		list, _ = biz.ContactsAndLastMessage(fromUser.UserID)
		for _, l := range list {
			if l.GroupID != testGroup.GroupID {
				continue
			}
			t.Log(l.LastMessage.Content)
			if l.LastMessage.Content != message.Content {
				t.Fatal(l.LastMessage)
			}
		}
	}
}

// go test -timeout 1h -run ^Test_aes$ github.com/ohko/chatroom -v -count=1
func Test_aes(t *testing.T) {
	txt := "hello"
	en, err := common.Encrypt([]byte(txt), []byte(config.AESKey))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(en))

	de, err := common.Decrypt(en, []byte(config.AESKey))
	if err != nil {
		t.Fatal(err)
	}

	if txt != string(de) {
		t.Fail()
	}
}
