package user

import (
	"fmt"
	"net/http"
	"strconv"

	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/mojiQAQ/dispatch/model"
)

type (
	ReqRegisterUser struct {
		*model.ReqBase
		WXID        string     `json:"wx_id"`
		PhoneNumber string     `json:"phone_number"`
		Role        model.Role `json:"role"`
	}

	RespRegisterUser struct {
		*model.RespBase

		Info *model.User `json:"info"`
	}

	ReqHandleBalance struct {
		*model.ReqBase
		Amount int64 `json:"amount"`
	}

	PrePayInfo struct {
		PrepayID  string `json:"prepay_id"`
		NonceStr  string `json:"nonce_str"`
		SignType  string `json:"sign_type"`
		PaySign   string `json:"pay_sign"`
		Package   string `json:"package"`
		Timestamp string `json:"timestamp"`
	}

	RespHandleBalanceRecharge struct {
		*model.RespBase
		PrepayInfo *PrePayInfo `json:"prepay_info"`
	}

	RespHandleBalanceWithdraw struct {
		*model.RespBase
	}

	ReqGetUserInfo struct {
		*model.ReqBase
	}

	RespGetUserInfo struct {
		*model.RespBase

		Info *model.User `json:"info"`
	}

	ReqGetUsers struct {
		*model.ReqBase
	}

	User struct {
		*model.TUser
		RoleCN string `json:"role_cn"`
	}

	RespGetUsers struct {
		*model.RespBase
		Users []*User `json:"users"`
	}

	ReqLogin struct {
		*model.ReqBase
	}

	RespLogin struct {
		*model.RespBase
		Token        string `json:"token"`
		OpenID       string `json:"openid"`
		IsRegistered bool   `json:"is_registered"`
	}

	ReqRegister struct {
		*model.ReqBase
		PhoneCode string     `json:"phone_code" valid:"required"`
		UserCode  string     `json:"user_code" valid:"required"`
		Role      model.Role `json:"role" valid:"required"`
	}

	RespRegister struct {
		*model.RespBase
		*model.User
	}

	Resource struct {
		Algorithm      string `json:"algorithm"`
		Ciphertext     string `json:"ciphertext"`
		AssociatedData string `json:"associated_data"`
		OriginalType   string `json:"original_type"`
		Nonce          string `json:"nonce"`
	}

	ReqPrepayCallback struct {
		*model.ReqBase
		ID           string   `json:"id"`
		CreateTime   string   `json:"create_time"`
		EventType    string   `json:"event_type"`
		ResourceType string   `json:"resource_type"`
		Resource     Resource `json:"resource"`
		Summary      string   `json:"summary"`
	}

	RespPrepayCallback struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	Credentials struct {
		TmpSecretID  string `json:"tmpSecretId,omitempty"`
		TmpSecretKey string `json:"tmpSecretKey,omitempty"`
		SessionToken string `json:"sessionToken,omitempty"`
	}

	RespGetTmpSecret struct {
		Credentials Credentials `json:"credentials"`
		ExpiredTime int         `json:"expiredTime,omitempty"`
		Expiration  string      `json:"expiration,omitempty"`
		StartTime   int         `json:"startTime,omitempty"`
	}

	ReqGetTrades struct {
		*model.ReqBase
	}

	Trade struct {
		*model.TTradeRecord
		TypeCN string `json:"type_cn"`
	}

	RespGetTrades struct {
		*model.RespBase
		Trades []*Trade `json:"trades"`
	}
)

func (c *Ctl) InitRouter(g *gin.RouterGroup) {

	g.POST("/register", c.HandleRegister)

	g.GET("/login", c.HandleLogin)

	// 查询所有用户
	g.GET("/users", c.HandleGetUsers)

	// 注册用户
	g.POST("/users", c.HandleRegisterUser)

	// 用户详情
	g.GET("/users/:openid", c.HandleGetUserInfo)

	// 交易记录
	g.GET("/users/transactions", c.HandleGetTransactions)

	// 充值
	g.POST("/users/balance/recharge", c.HandleBalanceRecharge)

	// 提现
	g.POST("/users/balance/withdraw", c.HandleBalanceWithdraw)

	// 微信支付确认回调
	g.POST("/wechat_prepay_callback", c.HandlePrepayCallback)

	g.GET("/tmpSecret", c.HandleGetTmpSecret)
}

func (c *Ctl) HandleRegister(ctx *gin.Context) {

	req := &ReqRegister{}
	err := ctx.ShouldBindBodyWith(req, binding.JSON)
	if err != nil {
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	ok, err := valid.ValidateStruct(req)
	if err != nil || !ok {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	user, err := c.Register(req.PhoneCode, req.UserCode, req.Role)
	if err != nil {
		c.Errorf("register failed, err=%s", err.Error())
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespRegister{
		RespBase: req.GenResponse(err),
		User:     user,
	})
}

func (c *Ctl) HandleLogin(ctx *gin.Context) {

	req := &ReqLogin{}
	code := ctx.Query("code")
	if len(code) == 0 {
		err := fmt.Errorf("not found login code")
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	role := ctx.Query("role")
	if len(role) == 0 {
		err := fmt.Errorf("not found role")
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	rd, err := strconv.Atoi(role)
	if rd == 0 {
		err := fmt.Errorf("invalid role")
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	user, err := c.Login(code, model.Role(rd))
	if err != nil {
		c.Errorf("login failed, err=%s", err.Error())
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespLogin{
		RespBase:     req.GenResponse(err),
		OpenID:       user.OpenID,
		IsRegistered: user.Phone == "",
	})
}

func (c *Ctl) HandleRegisterUser(ctx *gin.Context) {

	req := &ReqRegisterUser{}
	err := ctx.ShouldBindBodyWith(req, binding.JSON)
	if err != nil {
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	ok, err := valid.ValidateStruct(req)
	if err != nil || !ok {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	user, err := c.RegisterUser(req.WXID, req.PhoneNumber, req.Role)
	if err != nil {
		c.Errorf("register failed, err=%s", err.Error())
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespRegisterUser{
		RespBase: req.GenResponse(err),
		Info:     user,
	})
}

func (c *Ctl) HandleBalanceRecharge(ctx *gin.Context) {

	req := &ReqHandleBalance{}
	err := ctx.ShouldBindBodyWith(req, binding.JSON)
	if err != nil {
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	ok, err := valid.ValidateStruct(req)
	if err != nil || !ok {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	openid := ctx.Query("openid")
	if len(openid) == 0 {
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	user, err := c.GetUserByOpenID(openid)
	if err != nil {
		c.Errorf("check user failed: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	prepayInfo, err := c.DoManageBalance(user.ID, model.TypeRecharge, req.Amount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespHandleBalanceRecharge{
		RespBase:   req.GenResponse(nil),
		PrepayInfo: prepayInfo,
	})
}

func (c *Ctl) HandleBalanceWithdraw(ctx *gin.Context) {

	req := &ReqHandleBalance{}
	err := ctx.ShouldBindBodyWith(req, binding.JSON)
	if err != nil {
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	ok, err := valid.ValidateStruct(req)
	if err != nil || !ok {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	openid := ctx.Query("openid")
	if len(openid) == 0 {
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	user, err := c.GetUserByOpenID(openid)
	if err != nil {
		c.Errorf("check user failed: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	_, err = c.DoManageBalance(user.ID, model.TypeWithdraw, req.Amount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespHandleBalanceWithdraw{
		RespBase: req.GenResponse(nil),
	})
}

func (c *Ctl) HandleGetUserInfo(ctx *gin.Context) {

	req := &ReqGetUserInfo{}

	openID := ctx.Param("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid")
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
	}

	user, err := c.GetUserByOpenID(openID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespGetUserInfo{
		RespBase: req.GenResponse(err),
		Info:     user.User,
	})
}

func (c *Ctl) HandleGetUsers(ctx *gin.Context) {

	req := &ReqGetUsers{}

	users, err := c.GetUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	data := make([]*User, 0)
	for _, u := range users {
		data = append(data, &User{
			TUser:  u,
			RoleCN: model.RoleCN[u.Role],
		})
	}

	ctx.JSON(http.StatusOK, &RespGetUsers{
		RespBase: req.GenResponse(err),
		Users:    data,
	})
}

func (c *Ctl) HandlePrepayCallback(ctx *gin.Context) {

	err := c.PrepayCallback(ctx.Request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &RespPrepayCallback{
			Code:    "FAIL",
			Message: "失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, &RespPrepayCallback{
		Code:    "SUCCESS",
		Message: "成功",
	})
}

func (c *Ctl) HandleGetTmpSecret(ctx *gin.Context) {

	result, err := c.wx.GetTmpSecret()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &RespPrepayCallback{
			Code:    "FAIL",
			Message: "失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, &RespGetTmpSecret{
		Credentials: Credentials{
			TmpSecretID:  result.Credentials.TmpSecretID,
			TmpSecretKey: result.Credentials.TmpSecretKey,
			SessionToken: result.Credentials.SessionToken,
		},
		StartTime:   result.StartTime,
		ExpiredTime: result.ExpiredTime,
		Expiration:  result.Expiration,
	})
}

func (c *Ctl) HandleGetTransactions(ctx *gin.Context) {

	req := &ReqGetTrades{}
	ok, err := valid.ValidateStruct(req)
	if err != nil || !ok {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	openid := ctx.Query("openid")
	if len(openid) == 0 {
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	offset, limit := 0, 0
	offsetStr := ctx.Query("offset")
	offset, err = strconv.Atoi(offsetStr)
	if err != nil {
		c.Errorf("parsing offset failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	limitStr := ctx.Query("limit")
	limit, err = strconv.Atoi(limitStr)
	if err != nil {
		c.Errorf("parsing limit failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	user, err := c.GetUserByOpenID(openid)
	if err != nil {
		c.Errorf("check user failed: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	trades, err := c.trade.GetTradesPage(user.ID, offset, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	data := make([]*Trade, 0)
	for _, t := range trades {
		data = append(data, &Trade{
			TTradeRecord: t,
			TypeCN:       model.TradeTypeCN[t.Type],
		})
	}

	ctx.JSON(http.StatusOK, &RespGetTrades{
		RespBase: req.GenResponse(nil),
		Trades:   data,
	})
}
