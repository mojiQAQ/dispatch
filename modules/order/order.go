package order

import (
	"fmt"
	"github.com/mojiQAQ/dispatch/modules/utils"
	"gorm.io/gorm"
	"strings"

	"github.com/mojiQAQ/dispatch/model"
)

const (
	PublishOrderPrice  = 200 // 单位：分
	CompleteOrderPrice = 100
)

func (c *Ctl) GetOrders(states []string, userID, platform uint) ([]*model.TMasterOrder, error) {

	expr := make([]string, 0)
	if len(states) != 0 {
		expr = append(expr, fmt.Sprintf("state in (%s)", strings.Join(states, ",")))
	}

	if userID != 0 {
		expr = append(expr, fmt.Sprintf("user_id = %d", userID))
	}

	if platform != 0 {
		expr = append(expr, fmt.Sprintf("platform = %d", platform))
	}

	return c.GetMasterOrders(strings.Join(expr, " AND "))
}

func (c *Ctl) GetOrder(id uint) (*model.TMasterOrder, error) {

	order := &model.TMasterOrder{}
	err := c.db.Model(model.TMasterOrder{}).Where("id = ?", id).First(order).Error
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (c *Ctl) GetMasterOrders(condition string, args ...interface{}) ([]*model.TMasterOrder, error) {

	orders := make([]*model.TMasterOrder, 0)
	err := c.db.Model(model.TMasterOrder{}).Where(condition, args...).Order("id desc").Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// CreateMasterOrder 创建订单
func (c *Ctl) CreateMasterOrder(order *model.MasterOrder, openID string) (*model.TMasterOrder, error) {

	user, err := c.uc.GetUserByOpenID(openID)
	if err != nil {
		return nil, err
	}

	order.State = model.MOrderStateCreated
	order.UUID = utils.GenerateUUID()
	order.UserID = user.ID

	tOrder := &model.TMasterOrder{MasterOrder: order}
	err = c.db.Create(tOrder).Error
	if err != nil {
		return nil, err
	}

	return tOrder, nil
}

// ModifyMasterOrder 修改订单内容
func (c *Ctl) ModifyMasterOrder(id uint, order *model.MasterOrder) (*model.TMasterOrder, error) {

	oldOrder, err := c.GetOrder(id)
	if err != nil {
		return nil, err
	}

	// 仅初始化状态的订单可以修改订单内容
	if oldOrder.State != model.MOrderStateCreated {
		return nil, fmt.Errorf("not allow modify order")
	}

	tOrder := &model.TMasterOrder{MasterOrder: order}
	err = c.db.Model(model.TMasterOrder{}).Where("id = ?", id).Updates(tOrder).Error
	if err != nil {
		return nil, err
	}

	return tOrder, nil
}

// changeOrderState 修改订单状态
func (c *Ctl) changeOrderState(tx *gorm.DB, id uint, state model.OrderState) error {

	return tx.Model(model.TMasterOrder{}).Where("id = ?", id).Update("state", state).Error
}

// PayForMasterOrder 支付订单
func (c *Ctl) PayForMasterOrder(id uint) error {

	order, err := c.GetOrder(id)
	if err != nil {
		return err
	}

	// 仅已创建状态的订单可以被支付
	if order.State != model.MOrderStateCreated {
		return fmt.Errorf("only pay for order state is created")
	}

	// 开启事务
	tx := c.db.Begin()
	defer func() {
		if err != nil {
			rErr := tx.Rollback().Error
			if rErr != nil {
				c.Errorf("tx rollback failed, err=%v", rErr)
			}
		}
	}()

	// 支付订单
	err = c.uc.PayForPublishOrder(tx, order.UserID, order.Total*PublishOrderPrice, order.UUID)
	if err != nil {
		return err
	}

	// 更新订单状态为进行中
	err = c.changeOrderState(tx, id, model.MOrderStateDoing)
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

func (c *Ctl) PublishOrder(order *model.MasterOrder, openID string) (*model.TMasterOrder, error) {
	o, err := c.CreateMasterOrder(order, openID)
	if err != nil {
		return nil, err
	}

	err = c.PayForMasterOrder(o.ID)
	return o, err
}

func (c *Ctl) CreateSubOrder(mid uint, userID uint) (*model.TSubOrder, error) {

	user, err := c.uc.GetUser(userID)
	if err != nil {
		return nil, err
	}

	if user.Role != model.RoleWorker {
		return nil, fmt.Errorf("该用户不准接单")
	}

	mOrder, err := c.GetOrder(mid)
	if err != nil {
		return nil, err
	}

	if mOrder.State != model.MOrderStateDoing {
		return nil, fmt.Errorf("订单已失效")
	}

	subOrders, err := c.GetAllSubOrders(mid)
	if err != nil {
		return nil, err
	}

	if len(subOrders) >= int(mOrder.Total) {
		return nil, fmt.Errorf("订单已分配完成")
	}

	sOrder := &model.SubOrder{
		UUID:   utils.GenerateUUID(),
		MID:    mOrder.ID,
		UserID: userID,
		State:  model.SOrderStateAccept,
	}

	tSOrder := &model.TSubOrder{SubOrder: sOrder}
	err = c.db.Model(model.TSubOrder{}).Create(tSOrder).Error
	if err != nil {
		if strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
			return nil, fmt.Errorf("订单已接受")
		}
		return nil, err
	}

	return tSOrder, nil
}

func (c *Ctl) GetSubOrdersPlus(mid, userID uint, states []string) ([]*model.TSubOrder, error) {
	expr := make([]string, 0)
	args := make([]interface{}, 0)
	if userID != 0 {
		expr = append(expr, "user_id = ?")
		args = append(args, userID)
	}

	if mid != 0 {
		expr = append(expr, "mid = ?")
		args = append(args, mid)
	}

	if len(states) != 0 {
		expr = append(expr, "state in (?)")
		args = append(args, strings.Join(states, ","))
	}

	return c.getSubOrders(strings.Join(expr, " AND "), args...)
}

func (c *Ctl) getSubOrders(condition string, args ...interface{}) ([]*model.TSubOrder, error) {

	sOrders := make([]*model.TSubOrder, 0)
	err := c.db.Model(model.TSubOrder{}).Where(condition, args...).Find(&sOrders).Error
	if err != nil {
		return nil, err
	}

	return sOrders, nil
}

func (c *Ctl) GetAllSubOrders(mid uint) ([]*model.TSubOrder, error) {

	return c.getSubOrders("mid = ?", mid)
}

func (c *Ctl) GetSubOrderInfo(mid, sid uint) (*model.TSubOrder, error) {

	sOrder := &model.TSubOrder{}
	err := c.db.Model(model.TSubOrder{}).Where("id = ? AND mid = ?", sid, mid).First(sOrder).Error
	if err != nil {
		return nil, err
	}

	return sOrder, nil
}

func (c *Ctl) SubmitSubOrder(mid, sid uint, req *ReqSubmitSubOrders) error {

	// 查询子订单是否存在
	order, err := c.GetSubOrderInfo(mid, sid)
	if err != nil {
		return err
	}

	// 检查子订单是否允许提交
	if order.State != model.SOrderStateAccept && order.State != model.SOrderStateReject && order.State != model.SOrderStateSubmit {
		return fmt.Errorf("the sub order state only accept and reject can be submit")
	}

	// 提交订单
	order.State = model.SOrderStateSubmit
	order.Context = req.Context
	return c.db.Model(model.TSubOrder{}).Where("id = ?", sid).Updates(order).Error
}

func (c *Ctl) changeSubOrderState(tx *gorm.DB, sid uint, state model.OrderState) error {

	return tx.Model(model.TSubOrder{}).Where("id = ?", sid).
		Update("state", state).Error
}

// ApproveSubOrder 订单审核通过
func (c *Ctl) ApproveSubOrder(mid, sid uint) error {

	// 查询父订单
	masterOrder, err := c.GetOrder(mid)
	if err != nil {
		return err
	}

	if masterOrder.Complete == masterOrder.Total {
		return fmt.Errorf("master is complete")
	}

	// 查询子订单是否存在
	subOrder, err := c.GetSubOrderInfo(mid, sid)
	if err != nil {
		return err
	}

	// 开启事务
	tx := c.db.Begin()
	defer func() {
		if err != nil {
			rErr := tx.Rollback().Error
			if rErr != nil {
				c.Errorf("tx rollback failed, err=%v", rErr)
			}
		}
	}()

	// 修改订单状态
	err = c.changeSubOrderState(tx, sid, model.SOrderStateComplete)
	if err != nil {
	}

	err = tx.Model(model.TMasterOrder{}).Where("id = ?", mid).
		Update("complete", masterOrder.Complete+1).Error
	if err != nil {
		return err
	}

	// 支付佣金
	err = c.uc.RewardForOrder(tx, subOrder.UserID, CompleteOrderPrice, subOrder.UUID)
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

func (c *Ctl) ReviewSubOrder(mid, sid uint, auditorID uint, state model.OrderState) error {

	// 查询审核人信息
	auditor, err := c.uc.GetUser(auditorID)
	if err != nil {
		return err
	}

	// 仅审核员可以审核子订单
	if auditor.Role != model.RoleAuditor {
		return fmt.Errorf("only auditor can review sub order")
	}

	// 获取子订单信息
	subOrder, err := c.GetSubOrderInfo(mid, sid)
	if err != nil {
		return err
	}

	// 仅已提交的子订单可以被审核
	if subOrder.State != model.SOrderStateSubmit {
		return fmt.Errorf("only submit order can be review")
	}

	// 审核状态为完成，则通过子订单
	if state == model.SOrderStateComplete {
		return c.ApproveSubOrder(mid, sid)
	}

	// 审核状态为驳回，则驳回子订单
	if state == model.SOrderStateReject {
		return c.changeSubOrderState(c.db, sid, model.SOrderStateReject)
	}

	return nil
}
