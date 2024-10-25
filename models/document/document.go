package document

import "github.com/jinzhu/gorm"

// Document struct
type Document struct {
	gorm.Model
	LoanApplicationID int    `gorm:"not null"`
	DocumentType      string `gorm:"not null"`
	URL               string `gorm:"not null"`
}
