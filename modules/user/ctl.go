package user

import (
	"git.ucloudadmin.com/unetworks/app/pkg/log"
	"github.com/mojiQAQ/dispatch/model"
	"github.com/mojiQAQ/dispatch/modules/trade"
	"github.com/mojiQAQ/dispatch/modules/wechat"
	"gorm.io/gorm"
	"time"
)

type (
	Ctl struct {
		*log.Logger
		db *gorm.DB

		minBalance int64

		wx    *wechat.Ctl
		trade *trade.Ctl
	}
)

func NewCtl(logger *log.Logger, db *gorm.DB, t *trade.Ctl, w *wechat.Ctl) *Ctl {
	return &Ctl{
		Logger: logger,
		db:     db,

		minBalance: 2000,
		wx:         w,
		trade:      t,
	}
}

func (c *Ctl) Start() {

	ticker := time.NewTicker(time.Minute * 5)

	go func() {
		for {
			select {
			case <-ticker.C:
				c.Debugf("time to check order")
				// 检查超时未完成订单
				go c.checkTransferBatchResult()
			}
		}
	}()
}

// checkTransferBatchResult 检查转账记录
func (c *Ctl) autoCheckTransferBatchResult() {

	err := c.checkTransferBatchResult()
	if err != nil {
		c.Errorf("check wechat uncomplete transfer order failed, err=%s", err.Error())
		return
	}
}

func (c *Ctl) checkTransferBatchResult() error {
	records, err := c.trade.GetWxTransferRecord(model.WxTransferStateACCEPTED, model.WxTransferStatePROCESSING)
	if err != nil {
		return err
	}

	for _, r := range records {
		info, err := c.wx.CheckUnCompleteTransferOrder(r.TradeID)
		if err != nil {
			c.Errorf("check uncomplete transfer record failed, err=%s", err.Error())
			continue
		}

		user, err := c.GetUserByOpenID(r.OpenID)
		if err != nil {
			c.Errorf("get user failed, err=%s", err.Error())
			continue
		}

		err = c.UpdateTransferRecord(r.TradeID, *info.TransferBatch.BatchStatus,
			*info.TransferDetailList[0].DetailStatus, user, *info.TransferBatch.TotalAmount)
		if err != nil {
			c.Errorf("update withdraw trade record failed, err=%s", err.Error())
			continue
		}
	}

	return err
}
