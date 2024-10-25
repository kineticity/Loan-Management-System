package service

import (
	"errors"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"
	"time"

	"github.com/jinzhu/gorm"
)

type LogoutService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewLogoutService(db *gorm.DB, repository repository.Repository, log log.Logger) *LogoutService {
	return &LogoutService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}

// UpdateLoginInfo updates the logout time of the user's last login info entry
func (l *LogoutService) UpdateLoginInfo(user *user.User) error {
	// Transaction
	uow := repository.NewUnitOfWork(l.DB)
	defer uow.RollBack()

	// Fetch user by ID and preload LoginInfo
	user, err := l.repository.GetUserWithLoginInfo(uow, user.ID)
	if err != nil {
		return err
	}

	// Check if there is any LoginInfo data
	if len(user.LoginInfo) == 0 {
		return errors.New("no login info found for user")
	}

	// Update the last LoginInfo entry
	now := time.Now()
	logininfo := user.LoginInfo[len(user.LoginInfo)-1]

	// Ensure the primary key (ID) of the LoginInfo is set
	if logininfo.ID == 0 {
		return errors.New("invalid login info ID")
	}

	logininfo.LogoutTime = &now

	// Update logininfo in the database
	err = l.repository.Update(uow, logininfo)
	if err != nil {
		return err
	}

	// Update user data in the database (if needed)
	if err := l.repository.Update(uow, user); err != nil {
		return err
	}

	uow.Commit()
	return nil
}
