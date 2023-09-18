package user

import "github.com/mojiQAQ/dispatch/model"

func (c *Ctl) CreateUser(user *model.User) error {

	data := &model.TUser{
		User: user,
	}

	return c.db.Create(data).Error

}
