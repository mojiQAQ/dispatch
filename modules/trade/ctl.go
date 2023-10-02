package trade

import (
	"gorm.io/gorm"

	"github.com/mojiQAQ/dispatch/model"
)

type Ctl struct {
}

func NewCtl() *Ctl {

	return &Ctl{}
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
