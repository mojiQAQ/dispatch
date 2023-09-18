package model

import "gorm.io/gorm"

type (
	TUser struct {
		*gorm.Model
		*User
	}

	User struct {
		Role        Role    // 用户角色
		Balance     float64 // 余额
		PhoneNumber string  // 电话号码
		WxID        string  // 微信 ID
	}

	Role int64
)

const (
	RolePublisher     Role = iota + 1 // 发布人
	RoleWorker                        // 接单人
	RoleAuditor                       // 审核人
	RoleAdministrator                 // 管理员
)
