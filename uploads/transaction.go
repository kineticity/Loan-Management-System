package models

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Transaction struct {
	gorm.Model
	TransactionType        string    `json:"transaction_type"`
	Amount                 float64   `json:"amount"`
	Time                   time.Time `json:"time"`
	NewBalance             float64   `json:"new_balance"`
	CorrespondingAccountID uint      `json:"corresponding_account"` // Foreign key to Account
	AccountID              uint      `json:"account_id"`            // Foreign key to Account

}

func (t *Transaction) Create(db *gorm.DB) error {
	return db.Create(t).Error
}

func GetTransactionsByAccountID(db *gorm.DB, accountID int) ([]Transaction, error) {
	var transactions []Transaction
	if err := db.Where("account_id = ?", accountID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
