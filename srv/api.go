package srv

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/ohko/chatroom/biz"
	"github.com/ohko/chatroom/common"
	"github.com/ohko/chatroom/config"
)

type Api struct{}

func HandleApiFuncs(path string) {
	api := &Api{}
	t := reflect.TypeOf(api)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		fun := reflect.ValueOf(api).MethodByName(method.Name)
		http.HandleFunc(path+method.Name, common.Middleware(fun.Interface().(func(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData)))
	}
}

func (Api) Register(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		Account  string
		Password string
		RealName string
		Avatar   string
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	user := config.TableUser{
		Account:  data.Account,
		Password: data.Password,
		RealName: data.RealName,
		Avatar:   data.Avatar,
	}
	if err := biz.UserRegister(&user, r); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, user)
}

func (Api) UnRegister(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		Account  string
		Password string
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	info, err := biz.UserLogin(data.Account, data.Password, r)
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	if err := biz.UserDelete(info.UserID); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, "ok")
}

func (Api) Login(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		Account  string
		Password string
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	info, err := biz.UserLogin(data.Account, data.Password, r)
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	token, err := enToken(&Token{
		UserID: info.UserID,
	})
	if err != nil {
		return common.H_JSON(1, err.Error())
	}
	info.Token = token
	return common.H_JSON(0, info)
}

func (Api) UserUpdate(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		RealName string
		Avatar   string
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	token, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	info, err := biz.UserDetail(token.UserID)
	if err != nil {
		return common.H_JSON(1, err.Error())
	}
	info.RealName = data.RealName
	info.Avatar = data.Avatar
	info.UpdateTime = time.Now()

	if err := biz.UserUpdate(&info); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, info)
}

func (Api) UserDelete(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		UserID int
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	if err := biz.UserDelete(data.UserID); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, "ok")
}

func (Api) UserList(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	_, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	list, err := biz.UserList()
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, list)
}

func (Api) GroupCreate(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		GroupName string
		Users     []int
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	token, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	groupInfo := config.TableGroup{GroupName: data.GroupName, CreateUserID: token.UserID, OwnerID: token.UserID}
	if err := biz.GroupCreate(&groupInfo, data.Users); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, groupInfo)
}

func (Api) GroupUpdate(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		GroupID   int
		GroupName string
		Avatar    string
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	_, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	info, err := biz.GroupDetail(data.GroupID)
	if err != nil {
		return common.H_JSON(1, err.Error())
	}
	info.GroupName = data.GroupName
	info.Avatar = data.Avatar
	info.UpdateTime = time.Now()

	if err := biz.GroupUpdate(&info); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, info)
}

func (Api) GroupDelete(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		GroupID int
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	if err := biz.GroupDelete(data.GroupID); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, "ok")
}

func (Api) GroupJoinUsers(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		GroupID int
		Users   []int
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	_, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	if err := biz.UserGroupjoin(data.Users, data.GroupID); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, "ok")
}

func (Api) GroupRemoveUsers(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		GroupID int
		Users   []int
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	_, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	if err := biz.UserGroupRemove(data.Users, data.GroupID); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, "ok")
}

func (Api) GroupList(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	token, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	list, err := biz.UserGroupList(token.UserID)
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	ids := []int{}
	for _, l := range list {
		ids = append(ids, l.GroupID)
	}

	groups, err := biz.GroupListByID(ids)
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, groups)
}

func (Api) GroupListAll(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	_, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	groups, err := biz.GroupList()
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, groups)
}

func (Api) GroupUsers(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		GroupID int
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	_, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	list, err := biz.UserGroupListByGroupID(data.GroupID)
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, list)
}

func (Api) MessageSend(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		Type     string
		ToUserID int
		GroupID  int
		Content  string
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	token, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	msg := config.TableMessage{
		Type:       data.Type,
		FromUserID: token.UserID,
		ToUserID:   data.ToUserID,
		GroupID:    data.GroupID,
		Content:    data.Content,
	}

	if err = biz.MessageSend(&msg); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, msg)
}

func (Api) MessageList(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		GroupID    int
		FromUserID int
		PageSize   int
		Page       int
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	token, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	list, err := biz.MessageList(data.GroupID, data.FromUserID, token.UserID, (data.Page-1)*data.PageSize, data.PageSize)
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, list)
}

func (Api) MessageRead(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	var data struct {
		GroupID    int
		FromUserID int
	}
	if err := common.ReadPostData(ctx, r, &data); err != nil {
		return common.H_JSON(1, err.Error())
	}

	token, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	if err := biz.MessageRead(data.FromUserID, token.UserID, data.GroupID); err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, "ok")
}

func (Api) MessageLastList(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	token, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	list, err := biz.MessageLastList([]int{token.UserID})
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, list)
}

func (Api) ContactsAndLastMessage(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData {
	token, err := deToken(r.Header.Get(config.TokenName))
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	list, err := biz.ContactsAndLastMessage(token.UserID)
	if err != nil {
		return common.H_JSON(1, err.Error())
	}

	return common.H_JSON(0, list)
}

func GenerateToken(userID int) (string, error) {
	token := Token{
		UserID: userID,
	}
	return enToken(&token)
}

func enToken(token *Token) (string, error) {
	bs, _ := json.Marshal(token)
	result, err := common.Encrypt(bs, []byte(config.AESKey))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(result), nil
}

func deToken(token string) (*Token, error) {
	if token == "" {
		return nil, errors.New("need token")
	}
	var tk Token
	tmp, err := hex.DecodeString(token)
	if err != nil {
		return nil, err
	}
	de, err := common.Decrypt(tmp, []byte(config.AESKey))
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(de, &tk); err != nil {
		return nil, err
	}
	if tk.UserID == 0 {
		return nil, errors.New("login check error")
	}
	return &tk, nil
}

type Token struct {
	UserID int
}
