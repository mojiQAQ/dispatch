package model

import "gorm.io/gorm"

type (
	TTradeRecord struct {
		*gorm.Model
		TradeID string    // 交易 ID，如提现信息、充值信息
		UserID  uint32    // 用户 ID
		Type    TradeType // 交易类型
		Amount  float64   // 金额
	}

	TradeType uint32
)

const (
	TypeRecharge      TradeType = iota + 1 // 充值
	TypeWithdraw                           // 提现
	TypePublishOrder                       // 发布订单
	TypeCompleteOrder                      // 完成订单
)
