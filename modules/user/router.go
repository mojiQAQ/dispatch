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
		TradeType model.TradeType `json:"trade_type"`
		Amount    float64         `json:"amount"`
		TradeID   string          `json:"trade_id"`
	}

	RespHandleBalance struct {
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

	RespGetUsers struct {
		*model.RespBase

		Users []*model.TUser `json:"users"`
	}

	ReqLogin struct {
		*model.ReqBase
	}

	RespLogin struct {
		*model.RespBase
		Token  string `json:"token"`
		OpenID string `json:"openid"`
	}
)

func (c *Ctl) InitRouter(g *gin.RouterGroup) {

	g.GET("/login", c.HandleLogin)

	// 查询所有用户
	g.GET("/users", c.HandleGetUsers)

	// 注册用户
	g.POST("/users", c.HandleRegisterUser)

	// 用户详情
	g.GET("/users/:openid", c.HandleGetUserInfo)

	// 充值
	g.POST("/users/:openid/balance", c.HandleBalance)
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

	openID, err := c.Login(code, model.Role(rd))
	if err != nil {
		c.Errorf("login failed, err=%s", err.Error())
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespLogin{
		RespBase: req.GenResponse(err),
		OpenID:   openID,
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
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespRegisterUser{
		RespBase: req.GenResponse(err),
		Info:     user,
	})
}

func (c *Ctl) HandleBalance(ctx *gin.Context) {

	req := &ReqHandleBalance{}

	sid := ctx.Param("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
	}

	err = c.DoManageBalance(uint(id), req.TradeType, req.Amount, req.TradeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, req.GenResponse(nil))
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

	ctx.JSON(http.StatusOK, &RespGetUsers{
		RespBase: req.GenResponse(err),
		Users:    users,
	})
}
