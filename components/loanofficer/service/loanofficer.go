package service

import (
	"loanApp/app"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"

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
	// Validate newAdmin fields
	// if err := validateOfficer(officer); err != nil {
	// 	return err
	// }
	err := s.repository.Add(uow, &officer.User) // Add user first
	if err != nil {
		return err
	}

	// Then, use the same User ID to create the Admin
	officer.ID = officer.User.ID         // Link admin to the user by setting UserID
	err = s.repository.Add(uow, officer) // Add admin
	if err != nil {
		return err
	}

	if err := s.repository.Add(uow, officer); err != nil {
		return err
	}

	app.AllLoanOfficers = append(app.AllLoanOfficers, officer)
	uow.Commit()
	return nil
}

func (s *LoanOfficerService) GetAllLoanOfficers() ([]*user.LoanOfficer, error) {
	var officers []*user.LoanOfficer
	if err := s.repository.GetAll(nil, &officers); err != nil {
		return nil, err
	}
	return officers, nil
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

	uow.Commit()
	return nil
}
