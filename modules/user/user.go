package user

import (
	"errors"
	"github.com/mojiQAQ/dispatch/model"
	"gorm.io/gorm"
)

func (c *Ctl) CreateUser(user *model.User) error {

	data := &model.TUser{
		User: user,
	}

	return c.db.Create(data).Error

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

	user, err := c.GetUserByOpenID(auth.OpenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.RegisterUser(auth.OpenID, phone.PhoneNumber, role)
		}
		return nil, err
	}

	return user.User, nil
}

func (c *Ctl) UpdateUserInfo(openid, name, avatar string) (*model.User, error) {

	user, err := c.GetUserByOpenID(openid)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	if user.Name != name {
		data["name"] = name
	}

	if user.Avatar != avatar {
		data["avatar"] = avatar
	}

	err = c.db.Model(model.TUser{}).Where("openid = ?", openid).Updates(data).Error
	if err != nil {
		return nil, err
	}

	userNew, err := c.GetUserByOpenID(openid)
	if err != nil {
		return nil, err
	}

	return userNew.User, err
}
