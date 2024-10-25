package logininfo

import (
	"time"

	"github.com/jinzhu/gorm"
)

// LoginInfo struct
type LoginInfo struct {
	gorm.Model
	UserID     int       `gorm:"not null"`
	LoginTime  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	LogoutTime *time.Time `gorm:"nullable"`
}
