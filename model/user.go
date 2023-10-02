package model

import "gorm.io/gorm"

type (
	TUser struct {
		*gorm.Model
		*User
	}

	User struct {
		Role    Role    `gorm:"column:role" json:"role"`       // 用户角色
		Balance float64 `gorm:"column:balance" json:"balance"` // 余额
		Phone   string  `gorm:"column:phone" json:"phone"`     // 电话号码
		OpenID  string  `gorm:"column:openid" json:"openid"`   // 微信 OpenID
	}

	Role int
)

const (
	RolePublisher     Role = iota + 1 // 发布人
	RoleWorker                        // 接单人
	RoleAuditor                       // 审核人
	RoleAdministrator                 // 管理员
)
