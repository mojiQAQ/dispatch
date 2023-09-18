package order

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
	ReqGetMasterOrders struct {
		*model.ReqBase
	}

	RespGetMasterOrders struct {
		*model.RespBase
		Orders []*model.TMasterOrder `json:"orders"`
	}

	ReqCreateMasterOrder struct {
		*model.ReqBase
		*model.MasterOrder
	}

	RespCreateMasterOrder struct {
		*model.RespBase
		Order *model.TMasterOrder `json:"order"`
	}

	ReqGetMasterOrder struct {
		*model.ReqBase
	}

	RespGetMasterOrder struct {
		*model.RespBase
		Order *model.MasterOrder `json:"order"`
	}

	ReqModifyMasterOrder struct {
		*model.ReqBase
		*model.MasterOrder
	}

	RespModifyMasterOrder struct {
		*model.RespBase
		Order *model.TMasterOrder `json:"order"`
	}

	ReqOperateMasterOrder struct {
		*model.ReqBase
	}

	RespOperateMasterOrder struct {
		*model.RespBase
	}

	ReqCreateSubOrder struct {
		*model.ReqBase
		UserID uint32 `json:"user_id"`
	}

	RespCreateSubOrder struct {
		*model.RespBase
		SubOrder *model.TSubOrder `json:"sub_order"`
	}

	ReqGetSubOrders struct {
		*model.ReqBase
		State model.OrderState `json:"state"`
	}

	RespGetSubOrders struct {
		*model.RespBase
		SubOrders []*model.TSubOrder `json:"sub_orders"`
	}

	ReqGetSubOrderInfo struct {
		*model.ReqBase
	}

	RespGetSubOrderInfo struct {
		*model.RespBase
		SubOrder *model.TSubOrder `json:"sub_order"`
	}

	ReqSubmitSubOrders struct {
		*model.ReqBase
		UserID  uint32 `json:"user_id"`
		Context string `json:"context"`
	}

	RespSubmitSubOrders struct {
		*model.RespBase
	}

	ReqReviewSubOrders struct {
		*model.ReqBase
		AuditorID uint32           `json:"auditor_id"`
		State     model.OrderState `json:"state"`
	}

	RespReviewSubOrders struct {
		*model.RespBase
	}
)

func (c *Ctl) InitRouter(g *gin.RouterGroup) {

	// 查询订单
	g.GET("/orders", c.HandleGetOrders)

	// 创建订单
	g.POST("/orders", c.HandleCreateMasterOrder)

	// 查询订单详情
	g.GET("/orders/:id", c.HandleGetOrderInfo)

	// 修改订单
	g.PUT("/orders/:id", c.HandleModifyOrder)

	// 操作订单，action=submit/pay/publish(submit+pay)
	g.POST("/orders/:id", c.HandleOperateOrder)

	// 创建子订
	g.POST("/orders/:id/sub_orders", c.HandleCreateSubOrder)

	// 查询子订单列表
	g.GET("/orders/:id/sub_orders", c.HandleGetSubOrders)

	// 查询子订单详情
	g.GET("/orders/:id/sub_orders/:sid", c.HandleGetSubOrderInfo)

	// 提交子订单
	g.POST("/orders/:id/sub_orders/:sid", c.HandleSubmitSubOrder)

	// 修改子订单
	g.PUT("/orders/:id/sub_orders/:sid", c.HandleReviewSubOrder)

	// 审核子订单
	g.PUT("/orders/:id/sub_orders/:sid", c.HandleReviewSubOrder)
}

func (c *Ctl) HandleGetOrders(ctx *gin.Context) {

	req := &ReqGetMasterOrders{}

	ss := ctx.Param("state")
	sid := ctx.Param("user_id")
	sp := ctx.Param("platform")
	orders, err := c.GetOrders(ss, sid, sp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespGetMasterOrders{
		RespBase: req.GenResponse(err),
		Orders:   orders,
	})
}

func (c *Ctl) HandleCreateMasterOrder(ctx *gin.Context) {

	req := &ReqCreateMasterOrder{}
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

	order, err := c.CreateMasterOrder(req.MasterOrder)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespCreateMasterOrder{
		RespBase: req.GenResponse(err),
		Order:    order,
	})
}

func (c *Ctl) HandleGetOrderInfo(ctx *gin.Context) {

	req := &ReqGetMasterOrder{}

	sid := ctx.Query("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	order, err := c.GetOrder(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespGetMasterOrder{
		RespBase: req.GenResponse(err),
		Order:    order.MasterOrder,
	})
}

func (c *Ctl) HandleModifyOrder(ctx *gin.Context) {

	req := &ReqModifyMasterOrder{}
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

	sid := ctx.Query("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	order, err := c.ModifyMasterOrder(uint(id), req.MasterOrder)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespModifyMasterOrder{
		RespBase: req.GenResponse(err),
		Order:    order,
	})
}

func (c *Ctl) HandleOperateOrder(ctx *gin.Context) {

	req := &ReqOperateMasterOrder{}

	sid := ctx.Query("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	action := ctx.Param("action")
	switch action {
	case "submit":
		err = c.SubmitMasterOrder(uint(id))
	case "pay":
		err = c.PayForMasterOrder(uint(id))
	case "publish":
		err = c.PublishMasterOrder(uint(id))
	default:
		msg := "request params action invalid, only submit/pay/publish"
		c.Errorf(msg)
		ctx.JSON(http.StatusBadRequest, req.GenResponse(fmt.Errorf(msg)))
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespOperateMasterOrder{
		RespBase: req.GenResponse(err),
	})
}

func (c *Ctl) HandleCreateSubOrder(ctx *gin.Context) {

	req := &ReqCreateSubOrder{}
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

	sid := ctx.Query("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sOrder, err := c.CreateSubOrder(uint(id), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespCreateSubOrder{
		RespBase: req.GenResponse(err),
		SubOrder: sOrder,
	})
}

func (c *Ctl) HandleGetSubOrders(ctx *gin.Context) {

	req := &ReqGetSubOrders{}

	sid := ctx.Query("id")
	mid, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sOrders, err := c.GetSubOrders(uint(mid))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespGetSubOrders{
		RespBase:  req.GenResponse(err),
		SubOrders: sOrders,
	})
}

func (c *Ctl) HandleGetSubOrderInfo(ctx *gin.Context) {

	req := &ReqGetSubOrderInfo{}

	mid, err := strconv.Atoi(ctx.Query("mid"))
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sid, err := strconv.Atoi(ctx.Query("sid"))
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	order, err := c.GetSubOrderInfo(uint(mid), uint(sid))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespGetSubOrderInfo{
		RespBase: req.GenResponse(err),
		SubOrder: order,
	})
}

func (c *Ctl) HandleSubmitSubOrder(ctx *gin.Context) {

	req := &ReqSubmitSubOrders{}
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

	mid, err := strconv.Atoi(ctx.Query("mid"))
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sid, err := strconv.Atoi(ctx.Query("sid"))
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	err = c.SubmitSubOrder(uint(mid), uint(sid), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespSubmitSubOrders{
		RespBase: req.GenResponse(err),
	})
}

func (c *Ctl) HandleReviewSubOrder(ctx *gin.Context) {

	req := &ReqReviewSubOrders{}
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

	mid, err := strconv.Atoi(ctx.Query("mid"))
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sid, err := strconv.Atoi(ctx.Query("sid"))
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	err = c.ReviewSubOrder(uint(mid), uint(sid), req.AuditorID, req.State)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespReviewSubOrders{
		RespBase: req.GenResponse(err),
	})
}
