package loanapplication

import (
	"loanApp/models/document"
	"loanApp/models/installation"
	"time"

	"github.com/jinzhu/gorm"
)

type LoanApplication struct {
	gorm.Model
	LoanSchemeID    uint                         `gorm:"not null"` //int->uint //foreign key references loanschemes(id)
	CustomerID      uint                         `gorm:"not null"` //foreign key references customers(id)
	LoanOfficerID   uint                         `gorm:"not null"` //change back to not null later //foreign key references loanofficer(id)
	ApplicationDate time.Time                    `gorm:"default:CURRENT_TIMESTAMP"`
	DecisionDate    *time.Time                   `gorm:"nullable"`
	Status          string                       `gorm:"not null"`
	Amount          float64                      `gorm:"not null"`
	IsNPA           bool                         `gorm:"default:false"`
	Installations   []*installation.Installation `gorm:"foreignKey:LoanApplicationID"`
	Documents       []*document.Document         `gorm:"foreignKey:LoanApplicationID"`
}
