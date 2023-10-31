package user

import (
	"errors"
	"fmt"
	"github.com/mojiQAQ/dispatch/modules/wechat"
	"gorm.io/gorm"

	"git.ucloudadmin.com/unetworks/app/pkg/log"
	"github.com/mojiQAQ/dispatch/model"
	"github.com/mojiQAQ/dispatch/modules/trade"
)

type (
	Ctl struct {
		*log.Logger
		db *gorm.DB

		minBalance float64

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

func (c *Ctl) DoManageBalance(userID uint, tradeType model.TradeType, amount float64, tradeID string) error {

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

	switch tradeType {
	case model.TypeRecharge:
		return c.RechargeBalance(tx, userID, amount, tradeID)
	case model.TypeWithdraw:
		return c.WithdrawBalance(tx, userID, amount, tradeID)
	case model.TypePublishOrder:
		err = c.PayForPublishOrder(tx, userID, amount, tradeID)
	case model.TypeCompleteOrder:
		return c.RewardForOrder(tx, userID, amount, tradeID)
	}

	if err != nil {
		return err
	}

	return tx.Commit().Error
}

func (c *Ctl) RechargeBalance(tx *gorm.DB, userID uint, amount float64, tradeID string) error {

	user, err := c.GetUser(userID)
	if err != nil {
		return err
	}

	// 创建充值记录
	err = c.trade.AddTransactionRecord(tx, userID, model.TypeRecharge, amount, tradeID)
	if err != nil {
		return err
	}

	// 更新账户余额
	user.Balance = user.Balance + amount
	return tx.Model(model.TUser{}).Where("id = ?", userID).Updates(user).Error
}

func (c *Ctl) WithdrawBalance(tx *gorm.DB, userID uint, amount float64, tradeID string) error {

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
	err = c.trade.AddTransactionRecord(tx, userID, model.TypeWithdraw, amount, tradeID)
	if err != nil {
		return err
	}

	user.Balance = user.Balance - amount
	return tx.Model(model.TUser{}).Where("id = ?", userID).Updates(user).Error
}

func (c *Ctl) PayForPublishOrder(tx *gorm.DB, userID uint, amount float64, orderID string) error {

	user, err := c.GetUser(userID)
	if err != nil {
		return err
	}

	// 支付金额需要小于账户余额
	if amount > user.Balance {
		return fmt.Errorf("账户余额不足")
	}

	// 添加交易记录
	err = c.trade.AddTransactionRecord(tx, userID, model.TypePublishOrder, amount, orderID)
	if err != nil {
		return err
	}

	user.Balance = user.Balance - amount
	return tx.Model(model.TUser{}).Where("id = ?", userID).Updates(user).Error
}

func (c *Ctl) RewardForOrder(tx *gorm.DB, userID uint, amount float64, orderID string) error {

	user, err := c.GetUser(userID)
	if err != nil {
		return err
	}

	// 添加交易记录
	err = c.trade.AddTransactionRecord(tx, userID, model.TypeCompleteOrder, amount, orderID)
	if err != nil {
		return err
	}

	user.Balance = user.Balance + amount
	return tx.Model(model.TUser{}).Where("id = ?", userID).Updates(user).Error
}

func (c *Ctl) Login(code string, role model.Role) (string, error) {

	auth, err := c.wx.GetAuthKey(code)
	if err != nil {
		return "", err
	}

	userInfo, err := c.GetUserByOpenID(auth.OpenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user, err := c.RegisterUser(auth.OpenID, "", role)
			if err != nil {
				return "", err
			}

			return user.OpenID, err
		} else {
			return "", err
		}
	}

	return userInfo.OpenID, err
}

func (c *Ctl) GetPhoneNumber(code string) (string, error) {

	phone, err := c.wx.GetPhoneNumber(code)
	if err != nil {
		return "", nil
	}

	return phone.PurePhoneNumber, nil
}
