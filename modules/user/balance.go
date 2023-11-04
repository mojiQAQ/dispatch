package user

import (
	"fmt"
	"net/http"

	"gorm.io/gorm"

	"github.com/mojiQAQ/dispatch/model"
	"github.com/mojiQAQ/dispatch/modules/utils"
)

func (c *Ctl) DoManageBalance(userID uint, tradeType model.TradeType, amount int64) (*PrePayInfo, error) {

	var err error
	var prepayInfo *PrePayInfo
	tx := c.db.Begin()
	defer func() {
		if err != nil {
			rErr := tx.Rollback().Error
			if rErr != nil {
				c.Errorf("tx rollback failed, err=%v", rErr)
			}
		}
	}()

	tradeID := utils.GenerateUUID()
	switch tradeType {
	case model.TypeRecharge:
		prepayInfo, err = c.RechargeBalance(tx, userID, amount, tradeID)
	case model.TypeWithdraw:
		err = c.WithdrawBalance(tx, userID, amount, tradeID)
	default:
		err = fmt.Errorf("unsupport trade type: %v", model.TradeTypeCN[tradeType])
	}
	if err != nil {
		c.Errorf(err.Error())
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return prepayInfo, nil
}

func (c *Ctl) RechargeBalance(tx *gorm.DB, userID uint, amount int64, tradeID string) (*PrePayInfo, error) {

	user, err := c.GetUser(userID)
	if err != nil {
		return nil, err
	}

	resp, err := c.wx.CreateWechatPrePayOrder(user.OpenID, tradeID, fmt.Sprintf("余额充值-%d元", amount/100), amount)
	if err != nil {
		c.Errorf("create prepay order failed, err=%s", err.Error())
		return nil, err
	}

	// 创建微信支付记录
	err = c.trade.AddWxPayRecord(tx, user.OpenID, amount, tradeID, *resp.PrepayId)
	if err != nil {
		return nil, err
	}

	return &PrePayInfo{
		PrepayID:  *resp.PrepayId,
		NonceStr:  *resp.NonceStr,
		Package:   *resp.Package,
		SignType:  *resp.SignType,
		PaySign:   *resp.PaySign,
		Timestamp: *resp.TimeStamp,
	}, nil
}

func (c *Ctl) WithdrawBalance(tx *gorm.DB, userID uint, amount int64, tradeID string) error {

	user, err := c.GetUser(userID)
	if err != nil {
		return err
	}

	if amount > 20000 {
		return fmt.Errorf("最大提现 200")
	}

	// 提现金额需要大于账户余额
	if amount > user.Balance {
		return fmt.Errorf("超出账户余额")
	}

	// 如果是接单员，有最小余额限制
	if user.Role == model.RoleWorker {
		if (user.Balance - amount) < c.minBalance {
			return fmt.Errorf("需保持余额")
		}
	}

	resp, err := c.wx.TransferToWorker(user.OpenID, tradeID, fmt.Sprintf("余额提现-%d", amount), amount)
	if err != nil {
		return err
	}

	// 添加微信预转账交易记录
	err = c.trade.AddWxTransferRecord(tx, user.OpenID, amount, tradeID, *resp.BatchId)
	if err != nil {
		return err
	}

	return tx.Model(model.TUser{}).Where("id = ?", userID).Update("balance", user.Balance-amount).Error
}

func (c *Ctl) PayForPublishOrder(tx *gorm.DB, userID uint, amount int64, orderID string) error {

	user, err := c.GetUser(userID)
	if err != nil {
		return err
	}

	// 支付金额需要小于账户余额
	if amount > user.Balance {
		return fmt.Errorf("账户余额不足")
	}

	// 添加交易记录
	err = c.trade.AddTradeRecord(tx, userID, model.TypePublishOrder, amount, orderID)
	if err != nil {
		return err
	}

	user.Balance = user.Balance - amount
	return tx.Model(model.TUser{}).Where("id = ?", userID).Updates(user).Error
}

func (c *Ctl) ReturnUnCompleteOrder(tx *gorm.DB, userID uint, amount int64, orderID string) error {

	user, err := c.GetUser(userID)
	if err != nil {
		return err
	}

	// 添加退款交易记录
	err = c.trade.AddTradeRecord(tx, userID, model.TypeReturnOrder, amount, orderID)
	if err != nil {
		return err
	}

	return tx.Model(model.TUser{}).Where("id = ?", userID).Update("balance", user.Balance+amount).Error
}

func (c *Ctl) RewardForOrder(tx *gorm.DB, userID uint, amount int64, orderID string) error {

	user, err := c.GetUser(userID)
	if err != nil {
		return err
	}

	// 添加交易记录
	err = c.trade.AddTradeRecord(tx, userID, model.TypeCompleteOrder, amount, orderID)
	if err != nil {
		return err
	}

	user.Balance = user.Balance + amount
	return tx.Model(model.TUser{}).Where("id = ?", userID).Updates(user).Error
}

func (c *Ctl) PrepayCallback(req *http.Request) error {

	var err error
	tx := c.db.Begin()
	defer func() {
		if err != nil {
			rErr := tx.Rollback().Error
			if rErr != nil {
				c.Errorf("tx rollback failed, err=%v", rErr)
			}
		}
	}()

	transaction, err := c.wx.PrepayCallback(req)
	if err != nil {
		return err
	}

	user, err := c.GetUserByOpenID(*transaction.Payer.Openid)
	if err != nil {
		return err
	}

	err = c.trade.AddTradeRecord(tx, user.ID, model.TypeRecharge, *transaction.Amount.Total, *transaction.OutTradeNo)
	if err != nil {
		return err
	}

	err = c.trade.UpdateWxPayRecordState(tx, *transaction.OutTradeNo, *transaction.TradeState)
	if err != nil {
		return err
	}

	// 更新账户余额
	user.Balance = user.Balance + *transaction.Amount.Total
	err = tx.Model(model.TUser{}).Where("id = ?", user.ID).Updates(user).Error
	if err != nil {
		return err
	}

	return tx.Commit().Error
}

func (c *Ctl) UpdateTransferRecord(tradeID, batchStatus, detailStatus string, user *model.TUser, amount int64) error {

	var err error
	tx := c.db.Begin()
	defer func() {
		if err != nil {
			rErr := tx.Rollback().Error
			if rErr != nil {
				c.Errorf("tx rollback failed, err=%v", rErr)
			}
		}
	}()

	// 更新批量转账记录状态
	err = c.trade.UpdateWxTransferRecordState(c.db, tradeID, batchStatus)
	if err != nil {
		return err
	}

	// 成功的话，添加转账记录
	if detailStatus == "SUCCESS" {
		err = c.trade.AddTradeRecord(tx, user.ID, model.TypeWithdraw, amount, tradeID)
		if err != nil {
			return err
		}
	}

	// 转账失败，则退回用户余额
	if detailStatus == "FAIL" {
		err = tx.Model(model.TUser{}).Where("id = ?", user.ID).
			Update("balance = ?", user.Balance+amount).Error
		if err != nil {
			return err
		}
	}

	return tx.Commit().Error
}
