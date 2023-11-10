package trade

import (
	"gorm.io/gorm"
	"strings"

	"git.ucloudadmin.com/unetworks/app/pkg/log"

	"github.com/mojiQAQ/dispatch/model"
)

type Ctl struct {
	*log.Logger
	db *gorm.DB
}

func NewCtl(logger *log.Logger, db *gorm.DB) *Ctl {

	return &Ctl{logger, db}
}

func (c *Ctl) UpdateWxPayRecordState(db *gorm.DB, tradeID string, state string) error {
	return db.Model(model.TWxPayRecord{}).Where("trade_id = ?", tradeID).Update("state", state).Error
}

func (c *Ctl) AddWxPayRecord(db *gorm.DB, openid string, amount int64, tradeID, prepayID string) error {

	record := &model.TWxPayRecord{
		PrepayID: prepayID,
		TradeID:  tradeID,
		OpenID:   openid,
		Amount:   amount,
		State:    model.WxPayStatePrepay,
	}

	return db.Model(model.TWxPayRecord{}).Create(record).Error
}

func (c *Ctl) AddWxTransferRecord(db *gorm.DB, openid string, amount int64, tradeID, batchID string) error {

	record := &model.TWxTransferRecord{
		TradeID: tradeID,
		BatchID: batchID,
		OpenID:  openid,
		Amount:  amount,
		State:   model.WxTransferStateACCEPTED,
	}

	return db.Model(model.TWxPayRecord{}).Create(record).Error
}

func (c *Ctl) GetWxTransferRecord(states ...string) ([]*model.TWxTransferRecord, error) {

	rs := make([]*model.TWxTransferRecord, 0)
	err := c.db.Model(model.TWxTransferRecord{}).
		Where("state in ('')", strings.Join(states, "','")).Find(&rs).Error
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func (c *Ctl) UpdateWxTransferRecordState(db *gorm.DB, tradeID string, state string) error {
	return db.Model(model.TWxPayRecord{}).Where("trade_id = ?", tradeID).Update("state", state).Error
}

func (c *Ctl) AddTradeRecord(db *gorm.DB, userID uint, Type model.TradeType, amount, balance int64, TradeID string) error {

	record := &model.TTradeRecord{
		TradeID: TradeID,
		UserID:  userID,
		Type:    Type,
		Amount:  amount,
		Balance: balance,
	}

	return db.Model(model.TTradeRecord{}).Create(record).Error
}

func (c *Ctl) UpdateTradeRecordState(tx *gorm.DB, tradeID string, Type model.TradeType) error {

	return tx.Model(model.TTradeRecord{}).Where("trade_id = ?", tradeID).Update("type", Type).Error
}

func (c *Ctl) UpdateTradeRecordBalance(tx *gorm.DB, tradeID string, balance int64) error {

	return tx.Model(model.TTradeRecord{}).Where("trade_id = ?", tradeID).Update("balance", balance).Error
}

func (c *Ctl) getTrades(condition string, args ...interface{}) ([]*model.TTradeRecord, error) {

	trades := make([]*model.TTradeRecord, 0)
	err := c.db.Model(model.TTradeRecord{}).Where(condition, args...).Find(&trades).Error
	if err != nil {
		return nil, err
	}

	return trades, nil
}

func (c *Ctl) GetTrades(uuid string, userID int, tradeType model.TradeType) ([]*model.TTradeRecord, error) {

	expr := make([]string, 0)
	args := make([]interface{}, 0)
	if len(uuid) != 0 {
		expr = append(expr, "trade_id = ?")
		args = append(args, uuid)
	}

	if userID != 0 {
		expr = append(expr, "user_id = ?")
		args = append(args, userID)
	}

	if tradeType != 0 {
		expr = append(expr, "type = ?")
		args = append(args, tradeType)
	}

	return c.getTrades(strings.Join(expr, " AND "), args...)
}

func (c *Ctl) getTradesPage(offset, limit int, condition string, args ...interface{}) ([]*model.TTradeRecord, error) {

	trades := make([]*model.TTradeRecord, 0)
	err := c.db.Model(model.TTradeRecord{}).Where(condition, args...).
		Offset(offset).Limit(limit).Find(&trades).Error
	if err != nil {
		return nil, err
	}

	return trades, nil
}

func (c *Ctl) GetTradesPage(userID uint, offset, limit int) ([]*model.TTradeRecord, error) {

	expr := make([]string, 0)
	args := make([]interface{}, 0)

	if userID != 0 {
		expr = append(expr, "user_id = ?")
		args = append(args, userID)
	}

	return c.getTradesPage(offset, limit, strings.Join(expr, " AND "), args...)
}
