package service

import (
	"errors"
	"loanApp/app"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"
	"loanApp/utils/web"

	"strconv"

	"github.com/jinzhu/gorm"
)

type AdminService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewAdminService(db *gorm.DB, repository repository.Repository, log log.Logger) *AdminService {
	return &AdminService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}

// CreateAdmin creates a new admin in the database
func (u *AdminService) CreateAdmin(newAdmin *user.Admin) error {
	// Validate newAdmin fields
	if err := validateAdmin(newAdmin); err != nil {
		return err
	}

	// Transaction
	uow := repository.NewUnitOfWork(u.DB)
	defer uow.RollBack()

	err := u.repository.Add(uow, &newAdmin.User) // Add user first
	if err != nil {
		return err
	}

	// Then, use the same User ID to create the Admin
	newAdmin.ID = newAdmin.User.ID        // Link admin to the user by setting UserID
	err = u.repository.Add(uow, newAdmin) // Add admin
	if err != nil {
		return err
	}

	uow.Commit()
	app.AllAdmins = append(app.AllAdmins, newAdmin)
	return nil
}

// GetAllAdmins retrieves all admins from the database
func (u *AdminService) GetAllAdmins(allAdmins []*user.Admin, totalCount *int, parser web.Parser) error {
	uow := repository.NewUnitOfWork(u.DB)
	defer uow.RollBack()

	// Parse limit and offset as integers
	limit, err := strconv.Atoi(parser.Form.Get("limit"))
	if err != nil {
		limit = 12 // Set a default value if parsing fails
	}

	offset, err := strconv.Atoi(parser.Form.Get("offset"))
	if err != nil {
		offset = 0 // Set a default value if parsing fails
	}

	queryProcessors := []repository.QueryProcessor{
		u.repository.Filter("name=?", parser.Form.Get("name")),
		u.repository.Limit(limit),
		u.repository.Offset(offset),
	}

	if err := u.repository.GetAll(uow, allAdmins, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

// validateAdmin validates the admin data
func validateAdmin(admin *user.Admin) error {
	if admin.Name == "" {
		return errors.New("name cannot be empty")
	}
	if admin.Email == "" {
		return errors.New("email cannot be empty")
	}
	if admin.Password == "" {
		return errors.New("password cannot be empty")
	}
	// Additional validations can be added as needed
	return nil
}
