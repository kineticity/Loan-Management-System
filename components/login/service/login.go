package service

import (
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
func (l *LoginService) CreateLoginInfo(user *user.User,logininfo *logininfo.LoginInfo) error {

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
