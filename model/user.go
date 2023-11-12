package model

import "gorm.io/gorm"

type (
	TUser struct {
		*gorm.Model
		*User
	}

	User struct {
		Name    string `gorm:"column:name" json:"name"`       // 用户名
		Avatar  string `gorm:"column:avatar" json:"avatar"`   // 头像
		Role    Role   `gorm:"column:role" json:"role"`       // 用户角色
		Balance int64  `gorm:"column:balance" json:"balance"` // 余额，单位：分
		Phone   string `gorm:"column:phone" json:"phone"`     // 电话号码
		OpenID  string `gorm:"column:openid" json:"openid"`   // 微信 OpenID
		Credit  int    `gorm:"column:credit" json:"credit"`   // 信誉分
	}

	Role int
)

const (
	RolePublisher     Role = iota + 1 // 发布人
	RoleWorker                        // 接单人
	RoleAuditor                       // 审核人
	RoleAdministrator                 // 管理员
)

var RoleCN = map[Role]string{
	RolePublisher:     "发布人",
	RoleWorker:        "接单人",
	RoleAuditor:       "审核人",
	RoleAdministrator: "管理员",
}
