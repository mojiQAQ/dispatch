package model

import "gorm.io/gorm"

type (
	TTradeRecord struct {
		*gorm.Model
		TradeID string    `gorm:"column:trade_id" json:"trade_id"` // 交易 ID，如提现信息、充值信息
		UserID  uint      `gorm:"column:user_id" json:"user_id"`   // 用户 ID
		Type    TradeType `gorm:"column:type" json:"type"`         // 交易类型
		Amount  int64     `gorm:"column:amount" json:"amount"`     // 金额
		Balance int64     `gorm:"column:balance" json:"balance"`   // 余额
	}

	// TWxPayRecord 充值预支付记录
	TWxPayRecord struct {
		*gorm.Model
		PrepayID        string     `gorm:"column:prepay_id" json:"prepay_id"`                 // 预支付 ID
		TradeID         string     `gorm:"column:trade_id" json:"trade_id"`                   // 交易 ID，如提现信息、充值信息
		OpenID          string     `gorm:"column:openid" json:"openid"`                       // 用户 ID
		WxTransactionID string     `gorm:"column:wx_transaction_id" json:"wx_transaction_id"` // 微信交易 ID
		Amount          int64      `gorm:"amount" json:"amount"`                              // 金额，单位分
		State           WxPayState `gorm:"state" json:"state"`                                // 支付状态
	}

	TWxTransferRecord struct {
		*gorm.Model
		TradeID string     `gorm:"column:trade_id" json:"trade_id"` // 交易 ID，如提现信息、充值信息
		BatchID string     `gorm:"column:batch_id" json:"batch_id"` // 转账 ID
		OpenID  string     `gorm:"column:openid" json:"openid"`     // 用户 ID
		Amount  int64      `gorm:"amount" json:"amount"`            // 金额，单位分
		State   WxPayState `gorm:"state" json:"state"`              // 转账状态
	}

	TradeType  uint32
	WxPayState string
)

const (
	TypeRecharge      TradeType = iota + 1 // 充值
	TypeWithdraw                           // 提现
	TypePublishOrder                       // 商家发布订单
	TypeCompleteOrder                      // 用户完成订单
	TypeReturnOrder                        // 退费未完成订单
	TypeRecharging                         // 充值中
	TypeWithdrawing                        // 提现中
	TypeRechargeFail                       // 充值失败
	TypeWithdrawFail                       // 提现失败
)

var TradeTypeCN = map[TradeType]string{
	TypeRecharge:      "充值",
	TypeWithdraw:      "提现",
	TypePublishOrder:  "派单",
	TypeCompleteOrder: "接单",
	TypeReturnOrder:   "退费",
	TypeRecharging:    "充值中",
	TypeWithdrawing:   "提现中",
	TypeRechargeFail:  "充值失败",
	TypeWithdrawFail:  "提现失败",
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

const (
	WxTransferStateACCEPTED   = "ACCEPTED"
	WxTransferStatePROCESSING = "PROCESSING"
	WxTransferStateFINISHED   = "FINISHED"
	WxTransferStateCLOSED     = "CLOSED"

	WxTransferStateINIT = "INIT"
)
