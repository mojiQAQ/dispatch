package model

import "gorm.io/gorm"

type (
	TTradeRecord struct {
		*gorm.Model
		TradeID string    `gorm:"column:trade_id" json:"trade_id"` // 交易 ID，如提现信息、充值信息
		UserID  uint      `gorm:"column:user_id" json:"user_id"`   // 用户 ID
		Type    TradeType `gorm:"column:type" json:"type"`         // 交易类型
		Amount  float64   `gorm:"amount" json:"amount"`            // 金额
	}

	TWxPayRecord struct {
		*gorm.Model
		PrepayID        string     `gorm:"column:prepay_id" json:"prepay_id"` // 预支付 ID
		TradeID         string     `gorm:"column:trade_id" json:"trade_id"`   // 交易 ID，如提现信息、充值信息
		OpenID          string     `gorm:"column:openid" json:"openid"`
		WxTransactionID string     `gorm:"column:wx_transaction_id" json:"wx_transaction_id"`
		Amount          float64    `gorm:"amount" json:"amount"` // 金额
		State           WxPayState `gorm:"state" json:"state"`
	}

	TradeType  uint32
	WxPayState string
)

const (
	TypeRecharge      TradeType = iota + 1 // 充值
	TypeWithdraw                           // 提现
	TypePublishOrder                       // 发布订单
	TypeCompleteOrder                      // 完成订单
)

var TradeTypeCN = map[TradeType]string{
	TypeRecharge:      "充值",
	TypeWithdraw:      "提现",
	TypePublishOrder:  "派单",
	TypeCompleteOrder: "接单",
}

const (
	WxPayStatePrepay     = "Prepay"
	WxPayStateSUCCESS    = "SUCCESS"
	WxPayStateREFUND     = "REFUND"
	WxPayStateNOTPAY     = "NOTPAY"
	WxPayStateCLOSED     = "CLOSED"
	WxPayStateREVOKED    = "REVOKED"
	WxPayStateUSERPAYING = "USERPAYING"
	WxPayStatePAYERROR   = "PAYERROR"
)
