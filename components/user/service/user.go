package service

import (
	"errors"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
	// "golang.org/x/crypto/bcrypt"
)

type UserService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

// GetUserByEmail retrieves a user by their email address
func (u *UserService) GetUserByEmail(email string) (*user.User, error) {
	var foundUser user.User
	err := u.DB.Where("email = ?", email).First(&foundUser).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("user not found")
		}
		return nil, err // Return the underlying database error
	}
	return &foundUser, nil
}

// AuthenticateUser validates the user's credentials (email and password)
func (u *UserService) AuthenticateUser(email, password string) (*user.User, error) {
	foundUser, err := u.GetUserByEmail(email)
	if err != nil {
		return nil, err // If user is not found, return the error
	}

	// Verify password
	if foundUser.Password != password {
		return nil, errors.New("invalid password")
	}

	// If the user is inactive, return an error
	if !foundUser.IsActive {
		return nil, errors.New("user is inactive")
	}

	return foundUser, nil
}

// NewUserService creates a new UserService instance
func NewUserService(db *gorm.DB, repository repository.Repository, log log.Logger) *UserService {
	return &UserService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}
