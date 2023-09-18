package user

import (
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
)

func (c *Ctl) InitRouter(g *gin.RouterGroup) {

	// 查询所有用户
	g.GET("/users")

	// 注册用户
	g.POST("/users", c.HandleRegisterUser)

	// 用户详情
	g.GET("/users/:id", c.HandleGetUserInfo)

	// 充值
	g.POST("/users/:id/balance", c.HandleBalance)
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

	sid := ctx.DefaultQuery("id", "0")
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
	}

	err = c.DoManageBalance(uint32(id), req.TradeType, req.Amount, req.TradeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, req.GenResponse(nil))
}

func (c *Ctl) HandleGetUserInfo(ctx *gin.Context) {

	req := &ReqGetUserInfo{}

	sid := ctx.DefaultQuery("id", "0")
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
	}

	user, err := c.GetUser(uint32(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespGetUserInfo{
		RespBase: req.GenResponse(err),
		Info:     user,
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
