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
		Context  string     `gorm:"context" json:"context"`                      // 订单内容
		Remark   string     `gorm:"remark" json:"remark"`                        // 订单备注
		Platform Platform   `gorm:"platform" json:"platform" valid:"required"`   // 订单平台
		UserID   uint       `gorm:"user_id" json:"user_id"`                      // 创建人 ID
		State    OrderState `gorm:"state" json:"state"`                          // 订单状态
		Total    uint32     `gorm:"total" json:"total" valid:"required"`         // 总数量
		Complete uint32     `gorm:"complete" json:"complete"`                    // 已完成
		FinishAt time.Time  `gorm:"finish_at" json:"finish_at" valid:"required"` // 订单截止时间
	}

	SubOrder struct {
		*gorm.Model
		MID     uint       `gorm:"column:mid" json:"mid"`         // 关联父订单 ID
		UUID    string     `gorm:"column:uuid" json:"uuid"`       // 订单 UUID
		UserID  uint       `gorm:"column:user_id" json:"user_id"` // 创建人 ID
		State   OrderState `gorm:"column:state" json:"state"`     // 订单状态
		Context string     `gorm:"column:context" json:"context"` // 截图
	}

	OrderState int64
	Platform   int64
)

const (

	/*
				Created  ----> Cancel
				   |
				 Doing
			       |
			      / \
		      Done   Finish
	*/
	//
	MOrderStateCreated OrderState = iota + 1 // 已创建：此时订单可支付、可修改、可取消或超时未支付自动取消
	MOrderStateCancel                        // 取消：此时订单
	MOrderStateDoing                         // 进行中：此时订单已支付，接单员可以开始接单
	MOrderStateDone                          // 已完成：在订单截止时间所有接单人都已完成
	MOrderStateFinish                        // 已结束：在订单截止时间未全部完成

)

const (
	/*
			             Accept
						   |
			              / \
		    ｜------->Submit  Timeout
			｜			|
			｜		   / \
		     -----Reject   Complete
	*/
	SOrderStateAccept   OrderState = iota + 1 // 已接受
	SOrderStateSubmit                         // 已提交
	SOrderStateTimeout                        // 超时取消
	SOrderStateComplete                       // 已完成
	SOrderStateReject                         // 驳回
)

const (
	PlatformTB Platform = iota + 1 // 淘宝
	PlatformTM                     // 天猫
	PlatformJD                     // 京东
	PlatformDY                     // 抖音
	PlatformKS                     // 快手
)
