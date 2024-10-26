package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type User struct {
	gorm.Model
	Username   string     `gorm:"unique;not null"`
	Password   string     `gorm:"not null"`
	FirstName  string     `gorm:"not null"`
	LastName   string     `gorm:"not null"`
	IsAdmin    bool       `gorm:"default:false"`
	IsCustomer bool       `gorm:"default:false"`
	Accounts   []*Account `gorm:"foreignkey:CustomerID;references:ID"`
}

func (u *User) Create(db *gorm.DB) error {
	return db.Create(u).Error
}

func (u *User) Delete(db *gorm.DB) error {
	return db.Delete(u).Error
}

func GetUserByID(db *gorm.DB, id int) (*User, error) {
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetAllUsers(db *gorm.DB) ([]User, error) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (user *User) Update(tx *gorm.DB) error {
	return tx.Save(user).Error
}

func (u *User) AddAccount(db *gorm.DB, account *Account) error {
    account.CustomerID = u.ID



    if err := db.Model(u).Association("Accounts").Append(account).Error; err != nil {
        return err
    }

    return nil
}
