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

func (c *Ctl) AddTransactionRecord(db *gorm.DB, userID uint, Type model.TradeType, amount float64, TradeID string) error {

	record := &model.TTradeRecord{
		TradeID: TradeID,
		UserID:  userID,
		Type:    Type,
		Amount:  amount,
	}

	return db.Model(model.TTradeRecord{}).Create(record).Error
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
