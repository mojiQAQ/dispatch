package order

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"git.ucloudadmin.com/unetworks/app/pkg/log"
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

	ticker := time.NewTicker(time.Second * 10)

	go func() {
		for {
			select {
			case <-ticker.C:
				c.Debugf("time to check order")
				// 检查超时未完成订单
				go c.checkFinishOrder()
				go c.checkUnPayOrder()
				go c.checkAcceptOrder()
			}
		}
	}()
}

// checkUnPayOrder 检查未支付订单
func (c *Ctl) checkUnPayOrder() {

	orders, err := c.GetMasterOrders("state = ?", model.MOrderStateCreated)
	if err != nil {
		return
	}

	for _, order := range orders {
		// 如果订单处于待支付状态 10 分钟，则自动取消
		if order.State == model.MOrderStateCreated {
			if time.Now().Sub(order.UpdatedAt).Minutes() >= 10 {
				err = c.changeOrderState(c.db, order.ID, model.MOrderStateCancel)
				if err != nil {
					c.Errorf("cancel timeout order failed, uuid=%s, err=%v", order.UUID, err)
					continue
				}
			}
		}
	}
}

// checkFinishOrder 自动结束未完成订单
func (c *Ctl) checkFinishOrder() {

	orders, err := c.GetMasterOrders("state = ?", model.MOrderStateDoing)
	if err != nil {
		return
	}

	for _, order := range orders {
		c.Debugf("now: %v, finish: %v", time.Now(), order.FinishAt)
		if time.Now().After(order.FinishAt) {
			c.Infof("update finish order")
			err = c.AutoFinishMOrder(order)
			if err != nil {
				c.Errorf("auto finish order uuid=%s failed, err=%s", order.UUID, err.Error())
				continue
			}
		}
	}
}

// checkAcceptOrder 自动结束未完成子订单
func (c *Ctl) checkAcceptOrder() {

	subOrders, err := c.GetSubOrdersPlus(0, 0, []string{
		strconv.Itoa(int(model.SOrderStateAccept)),
		strconv.Itoa(int(model.SOrderStateReject)),
	})
	if err != nil {
		return
	}

	for _, so := range subOrders {
		if time.Now().Sub(so.CreatedAt).Minutes() >= 10 {
			err = c.AutoFinishSubOrder(so)
			if err != nil {
				c.Errorf("auto finish sub order uuid=%s failed, err=%s", so.UUID, err.Error())
				continue
			}
		}
	}
}
