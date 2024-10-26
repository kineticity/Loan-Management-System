package models

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Account struct {
	gorm.Model
	CustomerID uint           `gorm:"not null" json:"customer_id"` // Foreign Key to User table
	BankID     uint           `gorm:"not null" json:"bank_id"`     //Foreign key to Banks table
	Balance    float64        `json:"balance"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	Passbook   []*Transaction `gorm:"foreignKey:AccountID;refernces:ID" json:"passbook"`
}

func (a *Account) Create(db *gorm.DB) error {
	return db.Create(a).Error
}

func CheckIfAccountExists(tx *gorm.DB, customerID uint, bankID uint) error {
	var existingAccount Account

	// Check if account exists with the given customerID and bankID
	if err := tx.Where("customer_id = ? AND bank_id = ?", customerID, bankID).First(&existingAccount).Error; err == nil {
		// If no error, it means the account exists
		return errors.New("user already has an account in this bank")
	} else if err != gorm.ErrRecordNotFound {
		// If the error is something other than "record not found", return the error
		return err
	}

	// Return nil if no account exists
	return nil
}

//CheckIfAccountBelongsToCustomer verifies if an account exists and belongs to a specific customer
func CheckIfAccountBelongsToCustomer(tx *gorm.DB, accountID uint, customerID uint) (*Account, error) {
	var account Account

	// Query the account based on accountID and customerID
	if err := tx.Where("id = ? AND customer_id = ?", accountID, customerID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound // Return specific error if account is not found
		}
		return nil, err // Return other errors if any
	}

	return &account, nil
}

func (a *Account) Delete(db *gorm.DB) error {
	return db.Delete(a).Error
}

func GetAccountByID(db *gorm.DB, id int) (*Account, error) { //global account
	var account Account
	if err := db.First(&account, id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func GetAccountByIDAndCustomerID(tx *gorm.DB, accountID uint, customerID uint) (*Account, error) {
	var account Account

	// Query the account based on accountID and customerID
	if err := tx.Where("id = ? AND customer_id = ?", accountID, customerID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound // Return specific error if account is not found
		}
		return nil, err // Return other errors if any
	}

	return &account, nil
}


// func GetAllAccounts(db *gorm.DB) ([]Account, error) {
// 	var accounts []Account
// 	if err := db.Find(&accounts).Error; err != nil {
// 		return nil, err
// 	}
// 	return accounts, nil
// }

func GetAccountsByCustomerID(tx *gorm.DB, customerID uint) ([]Account, error) {
	var accounts []Account

	if err := tx.Where("customer_id = ?", customerID).Find(&accounts).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, errors.New("no accounts found for this customer")
	}

	return accounts, nil
}

func (a *Account) Update(db *gorm.DB) error {
	return db.Save(a).Error
}

func (a *Account) Withdraw(db *gorm.DB, amount float64) error {
	if a.Balance < amount {
		return errors.New("insufficient balance")
	}
	if (a.Balance - amount) < 2000 {
		return errors.New("minimum 2000 balance should be maintained")
	}
	a.Balance -= amount
	return db.Save(a).Error
}

func (a *Account) Deposit(db *gorm.DB, amount float64) error {
	a.Balance += amount
	return db.Save(a).Error
}

func (a *Account) Transfer(db *gorm.DB, toAccount *Account, amount float64) error {
	if a.Balance < amount {
		return errors.New("insufficient balance")
	}
	if (a.Balance - amount) < 2000 {
		return errors.New("minimum 2000 balance should be maintained")
	}

	a.Balance -= amount
	toAccount.Balance += amount

	if err := db.Save(a).Error; err != nil {
		return err
	}

	return db.Save(toAccount).Error
}

func (account *Account) AddToPassbook(tx *gorm.DB, transaction *Transaction) error {
	account.Passbook = append(account.Passbook, transaction)
	fmt.Println("account passbook:",account.Passbook,transaction)
	if err := tx.Save(account).Error; err != nil {
		return err
	}
	return nil
}

func (account *Account) PrintPassbook(tx *gorm.DB) error {
	var passbook []Transaction
	err := tx.Model(account).Association("Passbook").Find(&passbook).Error
	if err != nil {
		return err
	}

	fmt.Println("Passbook for Account ID:", account.ID)
	fmt.Println("-------------------------------------------------")
	for _, transaction := range passbook {
		fmt.Printf("Date: %s | Type: %s | Amount: %.2f | Balance: %.2f\n",
			transaction.Time.Format("2006-01-02 15:04:05"),
			transaction.TransactionType,
			transaction.Amount,
			transaction.NewBalance)
	}
	fmt.Println("-------------------------------------------------")
	return nil
}
