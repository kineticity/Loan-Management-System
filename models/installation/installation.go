package installation

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Installation struct
type Installation struct {
	gorm.Model
	LoanApplicationID int       `gorm:"not null"`
	AmountToBePaid    float64   `gorm:"not null"`
	DueDate           time.Time `gorm:"not null"`
	PaymentDate       time.Time `gorm:"nullable"`
	Status            string    `gorm:"not null"`
}
