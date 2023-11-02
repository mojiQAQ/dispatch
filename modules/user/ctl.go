package user

import (
	"fmt"
	"git.ucloudadmin.com/unetworks/app/pkg/log"
	"github.com/mojiQAQ/dispatch/model"
	"github.com/mojiQAQ/dispatch/modules/trade"
	"github.com/mojiQAQ/dispatch/modules/utils"
	"github.com/mojiQAQ/dispatch/modules/wechat"
	"gorm.io/gorm"
	"net/http"
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

		wx:    w,
		trade: t,
	}
}

func (c *Ctl) GetUsers() ([]*model.TUser, error) {

	users := make([]*model.TUser, 0)
	err := c.db.Model(model.TUser{}).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (c *Ctl) GetUser(id uint) (*model.User, error) {

	user := &model.TUser{}
	err := c.db.Model(model.TUser{}).Where("id = ?", id).First(user).Error
	if err != nil {
		return nil, err
	}

	return user.User, nil
}

func (c *Ctl) GetUserByOpenID(id string) (*model.TUser, error) {

	user := &model.TUser{}
	err := c.db.Model(model.TUser{}).Where("openid = ?", id).First(user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *Ctl) RegisterUser(openID, pn string, role model.Role) (*model.User, error) {

	user := &model.User{
		Role:    role,
		Balance: 0,
		Phone:   pn,
		OpenID:  openID,
	}

	err := c.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

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

	// 提现金额需要大于账户余额
	if amount > user.Balance {
		return fmt.Errorf("提现金额大于余额")
	}

	// 如果是接单员，有最小余额限制
	if user.Role == model.RoleWorker {
		if (user.Balance - amount) < c.minBalance {
			return fmt.Errorf("提现金额超出最大提现额度")
		}
	}

	// 添加交易记录
	err = c.trade.AddTradeRecord(tx, userID, model.TypeWithdraw, amount, tradeID)
	if err != nil {
		return err
	}

	user.Balance = user.Balance - amount
	return tx.Model(model.TUser{}).Where("id = ?", userID).Updates(user).Error
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

func (c *Ctl) Login(code string, role model.Role) (*model.User, error) {

	auth, err := c.wx.GetAuthKey(code, role)
	if err != nil {
		return nil, err
	}

	userInfo, err := c.GetUserByOpenID(auth.OpenID)
	if err != nil {
		//if errors.Is(err, gorm.ErrRecordNotFound) {
		//	user, err := c.RegisterUser(auth.OpenID, "", role)
		//	if err != nil {
		//		return nil, err
		//	}
		//
		//	return user, err
		//} else {
		//	return nil, err
		//}
		return nil, err
	}

	return userInfo.User, nil
}

func (c *Ctl) Register(phoneCode, userCode string, role model.Role) (*model.User, error) {

	// 获取手机号
	phone, err := c.wx.GetPhoneNumber(phoneCode, role)
	if err != nil {
		return nil, err
	}

	// 获取 OpenID
	auth, err := c.wx.GetAuthKey(userCode, role)
	if err != nil {
		return nil, err
	}

	return c.RegisterUser(auth.OpenID, phone.PhoneNumber, role)
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
