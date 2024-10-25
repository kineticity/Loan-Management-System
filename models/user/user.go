package user

import (
	"loanApp/models/loanapplication"
	"loanApp/models/logininfo"

	"github.com/jinzhu/gorm"
)

type Role string

// Base User struct
type User struct {
	gorm.Model
	Name      string                 `gorm:"not null"`
	Email     string                 `gorm:"unique;not null"`
	Password  string                 `gorm:"not null"`
	IsActive  bool                   `gorm:"default:true"`
	Role      Role                   `gorm:"not null"`
	LoginInfo []*logininfo.LoginInfo `gorm:"foreignKey:UserID"`
}

// Admin struct
type Admin struct {
	User
	LoanOfficers []*LoanOfficer `gorm:"foreignKey:AdminID"`
	// LoanSchemes  []*loanscheme.LoanScheme `gorm:"foreignKey:AdminID"`
}

// LoanOfficer struct
type LoanOfficer struct {
	User          // Embedding User struct
	CreatedBy     *Admin
	UpdatedBy     []*Admin
	AssignedLoans []*loanapplication.LoanApplication `gorm:"foreignKey:LoanOfficerID"`
}

// User struct for regular users
type Customer struct {
	User
	LoanApplications []*loanapplication.LoanApplication `gorm:"foreignKey:UserID"`
}
