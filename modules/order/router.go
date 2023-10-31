package order

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
		Orders []*Order `json:"orders"`
	}

	ReqGetAllMasterOrders struct {
		*model.ReqBase
	}

	Order struct {
		*model.TMasterOrder
		PlatformCN string `json:"platform_cn"`
		StateCN    string `json:"state_cn"`
		IsAccepted bool   `json:"is_accepted"`
		Accept     int    `json:"accept"`
		Review     int    `json:"review"`
	}

	RespGetAllMasterOrders struct {
		*model.RespBase
		Orders []*Order `json:"orders"`
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
		Order *Order `json:"order,omitempty"`
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
		SubOrders []*SubOrder `json:"sub_orders"`
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
		Context string `json:"context"`
	}

	RespSubmitSubOrders struct {
		*model.RespBase
	}

	ReqReviewSubOrders struct {
		*model.ReqBase
	}

	RespReviewSubOrders struct {
		*model.RespBase
	}

	ReqGetUserSubOrders struct {
		*model.ReqBase
	}

	SubOrder struct {
		*model.TSubOrder
		StateCN string `json:"state_cn"`
	}

	SubOrderExtend struct {
		*SubOrder
		MOrder *model.TMasterOrder `json:"m_order"`
	}

	RespGetUserSubOrders struct {
		*model.RespBase
		SubOrders []*SubOrderExtend `json:"sub_orders"`
	}
)

func (c *Ctl) InitRouter(g *gin.RouterGroup) {

	// 查询订单(商家用)
	g.GET("/orders", c.HandleGetOrders)

	// 查询全部订单(用户用)
	g.GET("/all_orders", c.HandleAllGetOrders)

	// 发布订单（创建并支付）
	g.POST("/orders", c.HandlePublishMasterOrder)

	// 查询订单详情
	g.GET("/orders/:id", c.HandleGetOrderInfo)

	// 修改订单
	g.PUT("/orders/:id", c.HandleModifyOrder)

	// 支付订单
	g.POST("/orders/:id", c.HandlePayOrder)

	// 创建子订单(接受订单)
	g.POST("/orders/:id/sub_orders", c.HandleCreateSubOrder)

	// 获取用户所有子订单(接单员)
	g.GET("/sub_orders", c.HandleGetUserSubOrders)

	// 查询子订单列表
	g.GET("/orders/:id/sub_orders", c.HandleGetSubOrders)

	// 查询子订单详情
	g.GET("/orders/:id/sub_orders/:sid", c.HandleGetSubOrderInfo)

	// 提交子订单
	g.POST("/orders/:id/sub_orders/:sid", c.HandleSubmitSubOrder)

	// 审核子订单
	g.PUT("/orders/:id/sub_orders/:sid", c.HandleReviewSubOrder)
}

func (c *Ctl) HandleGetOrders(ctx *gin.Context) {

	req := &ReqGetMasterOrders{}

	openid := ctx.Query("openid")
	if len(openid) == 0 {
		err := fmt.Errorf("invalid openid: %s", openid)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	user, err := c.uc.GetUserByOpenID(openid)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openid)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	states := make([]string, 0)
	state := ctx.Query("state")
	if len(state) != 0 {
		states = strings.Split(state, ",")
	}

	platform := ctx.DefaultQuery("platform", "0")
	pid, err := strconv.Atoi(platform)
	if err != nil {
		err := fmt.Errorf("invalid platform: %s", openid)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	userID := user.ID
	if user.Role == model.RoleAuditor || user.Role == model.RoleAdministrator {
		userID = 0
	}

	orders, err := c.GetOrders(states, userID, uint(pid))
	if err != nil {
		c.Errorf("get orders failed, err=%v", err)
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	sOrders, err := c.GetSubOrdersPlus(0, 0, []string{"1", "2"})
	if err != nil {
		c.Errorf("get sub orders failed, err=%v", err)
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	id2accept := make(map[uint]int)
	id2review := make(map[uint]int)

	for _, so := range sOrders {
		if _, exist := id2accept[so.ID]; !exist {
			id2accept[so.MID] = 0
		}

		if _, exist := id2review[so.ID]; !exist {
			id2review[so.MID] = 0
		}

		switch so.State {
		case model.SOrderStateAccept:
			id2accept[so.MID]++
		case model.SOrderStateSubmit:
			id2review[so.MID]++
		}
	}

	data := make([]*Order, 0)
	for _, o := range orders {
		data = append(data, &Order{
			TMasterOrder: o,
			PlatformCN:   model.PlatformCN[o.Platform],
			StateCN:      model.MOrderStateCN[o.State],
			Accept:       id2accept[o.ID],
			Review:       id2review[o.ID],
		})
	}

	ctx.JSON(http.StatusOK, &RespGetMasterOrders{
		RespBase: req.GenResponse(err),
		Orders:   data,
	})
}

func (c *Ctl) HandleAllGetOrders(ctx *gin.Context) {

	req := &ReqGetAllMasterOrders{}

	openid := ctx.Query("openid")
	if len(openid) == 0 {
		err := fmt.Errorf("invalid openid: %s", openid)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	user, err := c.uc.GetUserByOpenID(openid)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openid)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	states := make([]string, 0)
	state := ctx.Query("state")
	if len(state) != 0 {
		states = strings.Split(state, ",")
	}

	platform := ctx.DefaultQuery("platform", "0")
	pid, err := strconv.Atoi(platform)
	if err != nil {
		err := fmt.Errorf("invalid platform: %s", openid)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	orders, err := c.GetOrders(states, 0, uint(pid))
	if err != nil {
		c.Errorf("get orders failed, err=%v", err)
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	subOrders, err := c.GetSubOrdersPlus(0, user.ID, []string{})
	if err != nil {
		c.Errorf("get sub_orders failed, err=%v", err)
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	acceptOrder := make(map[uint]struct{})
	for _, subOrder := range subOrders {
		acceptOrder[subOrder.MID] = struct{}{}
	}

	allOrders := make([]*Order, 0)
	for _, order := range orders {
		newOrder := &Order{
			TMasterOrder: order,
			IsAccepted:   false,
		}

		if _, ok := acceptOrder[order.ID]; ok {
			newOrder.IsAccepted = true
		}

		allOrders = append(allOrders, newOrder)
	}

	ctx.JSON(http.StatusOK, &RespGetAllMasterOrders{
		RespBase: req.GenResponse(err),
		Orders:   allOrders,
	})
}

func (c *Ctl) HandlePublishMasterOrder(ctx *gin.Context) {

	req := &ReqCreateMasterOrder{}
	openID := ctx.Query("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

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

	order, err := c.PublishOrder(req.MasterOrder, openID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &RespCreateMasterOrder{
			RespBase: req.GenResponse(err),
			Order:    order,
		})
		return
	}

	ctx.JSON(http.StatusOK, &RespCreateMasterOrder{
		RespBase: req.GenResponse(err),
		Order:    order,
	})
}

func (c *Ctl) HandleGetOrderInfo(ctx *gin.Context) {

	req := &ReqGetMasterOrder{}
	openID := ctx.Query("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	user, err := c.uc.GetUserByOpenID(openID)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sid := ctx.Param("id")
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

	sOrders, err := c.GetSubOrdersPlus(order.ID, 0, []string{})
	if err != nil {
		c.Errorf("get sub_orders failed, err=%v", err)
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	isAccept := false
	accepted := 0
	review := 0
	for _, so := range sOrders {
		if so.UserID == user.ID {
			isAccept = true
		}

		switch so.State {
		case model.SOrderStateSubmit:
			review++
		case model.SOrderStateAccept:
			accepted++
		}
	}

	info := &Order{
		TMasterOrder: order,
		PlatformCN:   model.PlatformCN[order.Platform],
		StateCN:      model.MOrderStateCN[order.State],
		IsAccepted:   isAccept,
		Accept:       accepted,
		Review:       review,
	}

	ctx.JSON(http.StatusOK, &RespGetMasterOrder{
		RespBase: req.GenResponse(err),
		Order:    info,
	})
}

func (c *Ctl) HandleModifyOrder(ctx *gin.Context) {

	req := &ReqModifyMasterOrder{}
	openID := ctx.Query("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	_, err := c.uc.GetUserByOpenID(openID)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	err = ctx.ShouldBindBodyWith(req, binding.JSON)
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

	sid := ctx.Param("id")
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

func (c *Ctl) HandlePayOrder(ctx *gin.Context) {

	req := &ReqOperateMasterOrder{}
	openID := ctx.Query("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}
	_, err := c.uc.GetUserByOpenID(openID)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sid := ctx.Param("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	err = c.PayForMasterOrder(uint(id))
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
	openID := ctx.Query("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}
	user, err := c.uc.GetUserByOpenID(openID)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	err = ctx.ShouldBindBodyWith(req, binding.JSON)
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

	sid := ctx.Param("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sOrder, err := c.CreateSubOrder(uint(id), user.ID)
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
	openID := ctx.Query("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}
	_, err := c.uc.GetUserByOpenID(openID)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sid := ctx.Param("id")
	mid, err := strconv.Atoi(sid)
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sOrders, err := c.GetAllSubOrders(uint(mid))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	data := make([]*SubOrder, 0)
	for _, so := range sOrders {
		data = append(data, &SubOrder{
			TSubOrder: so,
			StateCN:   model.SOrderStateCN[so.State],
		})
	}

	ctx.JSON(http.StatusOK, &RespGetSubOrders{
		RespBase:  req.GenResponse(err),
		SubOrders: data,
	})
}

func (c *Ctl) HandleGetSubOrderInfo(ctx *gin.Context) {

	req := &ReqGetSubOrderInfo{}
	openID := ctx.Query("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}
	_, err := c.uc.GetUserByOpenID(openID)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	mid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sid, err := strconv.Atoi(ctx.Param("sid"))
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

	openID := ctx.Query("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}
	_, err = c.uc.GetUserByOpenID(openID)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	mid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sid, err := strconv.Atoi(ctx.Param("sid"))
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

	openID := ctx.Query("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}
	user, err := c.uc.GetUserByOpenID(openID)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	state := ctx.Query("state")
	if len(state) == 0 {
		err := fmt.Errorf("invalid state: %s", state)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	iState, err := strconv.Atoi(state)
	if err != nil {
		err := fmt.Errorf("invalid state: %s", state)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	mid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	sid, err := strconv.Atoi(ctx.Param("sid"))
	if err != nil {
		c.Errorf("request params invalid, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	err = c.ReviewSubOrder(uint(mid), uint(sid), user.ID, model.OrderState(iState))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, &RespReviewSubOrders{
		RespBase: req.GenResponse(err),
	})
}

func (c *Ctl) HandleGetUserSubOrders(ctx *gin.Context) {

	req := &ReqGetUserSubOrders{}
	openID := ctx.Query("openid")
	if len(openID) == 0 {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}
	user, err := c.uc.GetUserByOpenID(openID)
	if err != nil {
		err := fmt.Errorf("invalid openid: %s", openID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	states := make([]string, 0)
	state := ctx.Query("state")
	if len(state) != 0 {
		states = strings.Split(state, ",")
	}

	subOrders, err := c.GetSubOrdersPlus(0, user.ID, states)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
		return
	}

	sOrders := make([]*SubOrderExtend, 0)
	for _, s := range subOrders {

		mo, err := c.GetOrder(s.MID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, req.GenResponse(err))
			return
		}
		o := &SubOrderExtend{
			MOrder: mo,
			SubOrder: &SubOrder{
				TSubOrder: s,
			},
		}
		sOrders = append(sOrders, o)
	}

	ctx.JSON(http.StatusOK, &RespGetUserSubOrders{
		RespBase:  req.GenResponse(err),
		SubOrders: sOrders,
	})
}
