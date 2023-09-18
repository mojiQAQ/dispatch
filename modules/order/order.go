package order

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/mojiQAQ/dispatch/model"
)

func (c *Ctl) GetOrders(ss, sid, sp string) ([]*model.TMasterOrder, error) {

	state, err := strconv.Atoi(ss)
	if err != nil {
		return nil, err
	}

	userID, err := strconv.Atoi(sid)
	if err != nil {
		return nil, err
	}

	platform, err := strconv.Atoi(sp)
	if err != nil {
		return nil, err
	}

	expr := make([]string, 0)
	if state != 0 {
		expr = append(expr, fmt.Sprintf("state = %d", state))
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
	err := c.db.Model(model.TMasterOrder{}).Where(condition, args...).Find(orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// CreateMasterOrder 创建订单
func (c *Ctl) CreateMasterOrder(order *model.MasterOrder) (*model.TMasterOrder, error) {

	order.State = model.MOrderStateInit
	order.UUID = GenerateUUID()

	tOrder := &model.TMasterOrder{MasterOrder: order}
	err := c.db.Create(tOrder).Error
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
	if oldOrder.State != model.MOrderStateInit {
		return nil, fmt.Errorf("not allow modify order")
	}

	tOrder := &model.TMasterOrder{MasterOrder: order}
	err = c.db.Model(model.TMasterOrder{}).Where("id = ?", id).Updates(tOrder).Error
	if err != nil {
		return nil, err
	}

	return tOrder, nil
}

// chaneOrderState 修改订单状态
func (c *Ctl) chaneOrderState(tx *gorm.DB, id uint, state model.OrderState) error {

	return tx.Model(model.TMasterOrder{}).Where("id = ?", id).Update("state", state).Error
}

func (c *Ctl) SubmitMasterOrder(id uint) error {

	order, err := c.GetOrder(id)
	if err != nil {
		return err
	}

	// 仅初始化状态的订单可以被提交，提交后不可以修改
	if order.State != model.MOrderStateInit {
		return fmt.Errorf("only init order can be submit")
	}

	// 提交订单
	return c.chaneOrderState(c.db, id, model.MOrderStateCreated)
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
	err = c.uc.PayForPublishOrder(tx, order.UserID, float64(order.Total), order.UUID)
	if err != nil {
		return err
	}

	// 更新订单状态为进行中
	err = c.chaneOrderState(tx, id, model.MOrderStateDoing)
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// PublishMasterOrder 发布订单，包含提交订单和支付订单两步
func (c *Ctl) PublishMasterOrder(id uint) error {

	err := c.SubmitMasterOrder(id)
	if err != nil {
		return err
	}

	// 支付订单
	return c.PayForMasterOrder(id)
}

func (c *Ctl) CreateSubOrder(mid uint, req *ReqCreateSubOrder) (*model.TSubOrder, error) {

	mOrder, err := c.GetOrder(mid)
	if err != nil {
		return nil, err
	}

	if mOrder.State != model.MOrderStateDoing {
		return nil, fmt.Errorf("only ")
	}

	subOrders, err := c.GetSubOrders(mid)
	if err != nil {
		return nil, err
	}

	if len(subOrders) >= int(mOrder.Total) {
		return nil, fmt.Errorf("sub order is enough")
	}

	sOrder := &model.SubOrder{
		UUID:     GenerateUUID(),
		MID:      mOrder.UUID,
		UserID:   req.UserID,
		State:    model.SOrderStateAccept,
		FinishAt: time.Time{},
	}

	tSOrder := &model.TSubOrder{SubOrder: sOrder}
	err = c.db.Model(model.TSubOrder{}).Create(tSOrder).Error
	if err != nil {
		return nil, err
	}

	return tSOrder, nil
}

func (c *Ctl) GetSubOrders(mid uint) ([]*model.TSubOrder, error) {

	sOrders := make([]*model.TSubOrder, 0)
	err := c.db.Model(model.TSubOrder{}).Where("m_order_id = ?", mid).Find(sOrders).Error
	if err != nil {
		return nil, err
	}

	return sOrders, nil
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
	if order.State != model.SOrderStateAccept && order.State != model.SOrderStateReject {
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
	err = c.uc.RewardForOrder(tx, subOrder.UserID, 1, subOrder.UUID)
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

func (c *Ctl) ReviewSubOrder(mid, sid uint, auditorID uint32, state model.OrderState) error {

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
