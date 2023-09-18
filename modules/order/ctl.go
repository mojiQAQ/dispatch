package order

import (
	"time"

	"git.ucloudadmin.com/unetworks/app/pkg/log"
	"gorm.io/gorm"

	"github.com/mojiQAQ/dispatch/model"
	"github.com/mojiQAQ/dispatch/modules/user"
)

type Ctl struct {
	*log.Logger
	db *gorm.DB

	uc *user.Ctl
}

func NewCtl(logger *log.Logger, db *gorm.DB, uc *user.Ctl) *Ctl {
	return &Ctl{
		Logger: logger,
		db:     db,

		uc: uc,
	}
}

func (c *Ctl) Start() {

	ticker := time.NewTimer(time.Second * 30)

	for {
		select {
		case <-ticker.C:
			c.checkOrder()
		}
	}
}

func (c *Ctl) checkOrder() {

	orders, err := c.GetMasterOrders("state = ?", model.MOrderStateCreated)
	if err != nil {
		return
	}

	for _, order := range orders {
		// 如果订单处于待支付状态 10 分钟，则自动取消
		if order.State == model.MOrderStateCreated {
			if time.Now().Sub(order.UpdatedAt).Minutes() >= 10 {
				err = c.chaneOrderState(c.db, order.ID, model.MOrderStateCancel)
				if err != nil {
					c.Errorf("cancel timeout order failed, uuid=%s, err=%v", order.UUID, err)
					continue
				}
			}
		}

		subOrders, err := c.GetSubOrders(order.ID)
		if err != nil {
			return
		}

		for _, so := range subOrders {
			if so.State == model.SOrderStateAccept {
				if time.Now().Sub(so.CreatedAt).Minutes() >= 10 {
					err = c.changeSubOrderState(c.db, so.ID, model.SOrderStateTimeout)
					if err != nil {
						c.Errorf("set timeout sub order failed, uuid=%s, err=%v", order.UUID, err)
						continue
					}
				}
			}
		}
	}
}
