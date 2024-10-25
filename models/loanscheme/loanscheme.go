package loanscheme

import (
	"loanApp/models/loanapplication"
	"loanApp/models/user"

	"github.com/jinzhu/gorm"
)

// LoanScheme struct
type LoanScheme struct {
	gorm.Model
	Name                 string                             `gorm:"not null"`
	Category             string                             `gorm:"not null"` // retail or corporate
	CreatedBy            *user.Admin                        `gorm:"not null"`
	UpdatedBy            []*user.Admin                      `gorm:"not null"`
	InterestRate         float64                            `gorm:"not null"`
	Tenure               int                                `gorm:"not null"` //--???
	ApplicationsOfScheme []*loanapplication.LoanApplication `gorm:"foreignKey:LoanSchemeID"`
}