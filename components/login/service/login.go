package service

import (
	"errors"
	"loanApp/models/logininfo"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type LoginService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewLoginService(db *gorm.DB, repository repository.Repository, log log.Logger) *LoginService {
	return &LoginService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}

// CreateAdmin creates a new admin in the database
func (l *LoginService) CreateLoginInfo(user *user.User, logininfo *logininfo.LoginInfo) error {

	user.LoginInfo = append(user.LoginInfo, logininfo)

	// Transaction
	uow := repository.NewUnitOfWork(l.DB)
	defer uow.RollBack()

	err := l.repository.Add(uow, logininfo)
	if err != nil {
		return err
	}

	if err := l.repository.Update(uow, user); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

// GetActiveLoginInfo checks for an existing active login session for a user by user ID
func (ls *LoginService) GetActiveLoginInfo(userID int) (*logininfo.LoginInfo, error) {
	var loginInfo logininfo.LoginInfo

	// Query for an active session with the specified user ID
	err := ls.DB.Where("user_id = ? AND logout_time IS NULL", userID).First(&loginInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No active session found
		}
		return nil, err // Return other database errors if any
	}

	return &loginInfo, nil // Return the active login info if found
}
