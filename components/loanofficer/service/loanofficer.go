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

type LoanOfficerService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewLoanOfficerService(db *gorm.DB, repository repository.Repository, log log.Logger) *LoanOfficerService {
	return &LoanOfficerService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}

func (s *LoanOfficerService) CreateLoanOfficer(officer *user.LoanOfficer) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Validate newOfficer fields
	if err := validateLoanOfficer(officer); err != nil {
		return err
	}

	// First, add the User
	err := s.repository.Add(uow, &officer.User)
	if err != nil {
		return err
	}

	// Set the LoanOfficer ID to match the User ID
	officer.ID = officer.User.ID

	// Then, add the LoanOfficer
	err = s.repository.Add(uow, officer)
	if err != nil {
		return err
	}

	// Commit the transaction
	uow.Commit()

	// Append to the global list of Loan Officers
	app.AllLoanOfficers = append(app.AllLoanOfficers, officer)
	return nil
}

// GetAllAdmins retrieves all admins from the database
func (s *LoanOfficerService) GetAllLoanOfficers(allOfficers *[]*user.LoanOfficer, totalCount *int, parser web.Parser) error {
	uow := repository.NewUnitOfWork(s.DB)
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
		// s.repository.Filter("name=?", parser.Form.Get("name")),
		s.repository.Preload("LoginInfo"),
		// s.repository.Preload("UpdatedBy"),
		// s.repository.Preload("AssignedLoans"),
		s.repository.Limit(limit),
		s.repository.Offset(offset),
	}

	if err := s.repository.GetAll(uow, allOfficers, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanOfficerService) UpdateLoanOfficer(id string, updatedOfficer *user.LoanOfficer) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var officer user.LoanOfficer
	if err := s.repository.GetByID(uow, &officer, id); err != nil {
		return err
	}

	officer.Name = updatedOfficer.Name
	officer.Email = updatedOfficer.Email
	officer.Password = updatedOfficer.Password

	if err := s.repository.Update(uow, &officer); err != nil {
		return err
	}
	if err := s.repository.Update(uow, &officer.User); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanOfficerService) DeleteLoanOfficer(id string) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var officer user.LoanOfficer
	if err := s.repository.GetByID(uow, &officer, id); err != nil {
		return err
	}

	if err := s.repository.DeleteByID(uow, &officer, id); err != nil {
		return err
	}

	var user user.User
	if err := s.repository.GetByID(uow, &user, id); err != nil {
		return err
	}

	if err := s.repository.DeleteByID(uow, &user, id); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func validateLoanOfficer(loanOfficer *user.LoanOfficer) error {
	if loanOfficer.Name == "" {
		return errors.New("name cannot be empty")
	}
	if loanOfficer.Email == "" {
		return errors.New("email cannot be empty")
	}
	if loanOfficer.Password == "" {
		return errors.New("password cannot be empty")
	}
	// Additional validations can be added as needed
	return nil
}
