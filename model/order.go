package model

import (
	"time"

	"gorm.io/gorm"
)

type (
	TMasterOrder struct {
		*gorm.Model
		*MasterOrder
	}

	TSubOrder struct {
		*gorm.Model
		*SubOrder
	}

	MasterOrder struct {
		UUID     string     `gorm:"uuid" json:"uuid"`                            // 订单 UUID
		Name     string     `gorm:"name" json:"name" valid:"required"`           // 订单名称
		Platform Platform   `gorm:"platform" json:"platform" valid:"required"`   // 订单平台
		UserID   uint32     `gorm:"user_id" json:"user_id" valid:"required"`     // 创建人 ID
		State    OrderState `gorm:"state" json:"state"`                          // 订单状态
		Total    uint32     `gorm:"total" json:"total" valid:"required"`         // 总数量
		Complete uint32     `gorm:"complete" json:"complete"`                    // 已完成
		FinishAt time.Time  `gorm:"finish_at" json:"finish_at" valid:"required"` // 订单截止时间
		Context  string     `gorm:"context" json:"context"`                      // 订单内容
	}

	SubOrder struct {
		*gorm.Model
		UUID     string     `gorm:"uuid"`      // 订单 UUID
		MID      string     `gorm:"mid"`       // 关联父订单 ID
		UserID   uint32     `gorm:"user_id"`   // 创建人 ID
		State    OrderState `gorm:"state"`     // 订单状态
		FinishAt time.Time  `gorm:"finish_at"` // 订单截止时间
		Context  string     `gorm:"context"`
	}

	OrderState int64
	Platform   int64
)

const (

	/*
				  Init
				   |
				Created-------Cancel
				   |
				 Doing
			       |
			      / \
		      Done   Finish
	*/
	//
	MOrderStateInit    OrderState = iota + 1 // 初始化：已创建未支付，此时可以修改、取消订单
	MOrderStateCreated                       // 已创建：此时进入待支付状态且不可修改，可以选择取消订单或超时未支付则自动取消订单
	MOrderStateCancel                        // 取消：此时订单
	MOrderStateDoing                         // 进行中：此时订单已支付，接单员可以开始接单
	MOrderStateDone                          // 已完成：在订单截止时间所有接单人都已完成
	MOrderStateFinish                        // 已结束：在订单截止时间未全部完成

	/*
			             Accept
						   |
			              / \
		    ｜------->Submit  Timeout
			｜			|
			｜		   / \
		     -----Reject   Complete
	*/

	SOrderStateAccept   OrderState = iota + 10 // 已接受
	SOrderStateSubmit                          // 已提交
	SOrderStateTimeout                         // 超时取消
	SOrderStateComplete                        // 已完成
	SOrderStateReject                          // 驳回

	PlatformTB Platform = iota + 1 // 淘宝
	PlatformTM                     // 天猫
	PlatformJD                     // 京东
	PlatformDY                     // 抖音
	PlatformKS                     // 快手
)
