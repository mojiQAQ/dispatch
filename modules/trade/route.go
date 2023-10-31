package trade

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mojiQAQ/dispatch/model"
	"net/http"
	"strconv"
)

type (
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

	g.GET("/trades", c.HandleGetTrades)
}

func (c *Ctl) HandleGetTrades(ctx *gin.Context) {

	req := &ReqGetTrades{}
	aUUID := ctx.Param("uuid")
	userID := ctx.Param("user_id")
	if len(userID) == 0 {
		userID = "0"
	}

	uid, err := strconv.Atoi(userID)
	if err != nil {
		err := fmt.Errorf("invalid user_id: [%v]", userID)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	t := ctx.Param("type")
	if len(t) == 0 {
		t = "0"
	}

	iType, err := strconv.Atoi(t)
	if err != nil {
		err := fmt.Errorf("invalid type: [%v]", t)
		c.Errorf("parsing request failed, err=%s", err.Error())
		ctx.JSON(http.StatusBadRequest, req.GenResponse(err))
		return
	}

	trades, err := c.GetTrades(aUUID, uid, model.TradeType(iType))
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
		RespBase: req.GenResponse(err),
		Trades:   data,
	})
}
