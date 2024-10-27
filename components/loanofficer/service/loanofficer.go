package service

import (
	"errors"
	"fmt"
	"loanApp/app"
	"loanApp/models/installation"
	"loanApp/models/loanapplication"
	"loanApp/models/loanscheme"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"
	"loanApp/utils/web"
	"math"
	"strconv"
	"time"

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

	if err := validateLoanOfficer(officer); err != nil {
		return err
	}

	err := s.repository.Add(uow, &officer.User)
	if err != nil {
		return err
	}

	officer.ID = officer.User.ID

	err = s.repository.Add(uow, officer)
	if err != nil {
		return err
	}

	uow.Commit()

	app.AllLoanOfficers = append(app.AllLoanOfficers, officer)
	return nil
}

func (s *LoanOfficerService) GetAllLoanOfficers(allOfficers *[]*user.LoanOfficer, totalCount *int, parser web.Parser) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	limit, err := strconv.Atoi(parser.Form.Get("limit"))
	if err != nil {
		limit = 12 // default
	}

	offset, err := strconv.Atoi(parser.Form.Get("offset"))
	if err != nil {
		offset = 0 // default
	}

	queryProcessors := []repository.QueryProcessor{
		// s.repository.Filter("name=?", parser.Form.Get("name")),
		s.repository.Preload("LoginInfo"),
		// s.repository.Preload("UpdatedBy"),
		s.repository.Preload("AssignedLoans"), //uncomment
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
	return nil
}

// returns the loan officer with the least workload / assignedloanapplications
func (s *LoanOfficerService) GetLeastLoadedOfficer() (*user.LoanOfficer, error) {
	var officer user.LoanOfficer
	err := s.DB.Model(&user.LoanOfficer{}).
		Joins("LEFT JOIN loan_applications ON loan_officers.id = loan_applications.loan_officer_id").
		Group("loan_officers.id").
		Order("COUNT(loan_applications.id) ASC").
		First(&officer).Error

	if err != nil {
		return nil, err
	}
	return &officer, nil
}

// //by id get
func (s *LoanOfficerService) GetAssignedLoanApplications(loanOfficerID uint) ([]*loanapplication.LoanApplication, error) {
	var applications []*loanapplication.LoanApplication
	uow := repository.NewUnitOfWork(s.DB)
	// QueryProcessors to filter by LoanOfficerID and preload associations
	queryProcessors := []repository.QueryProcessor{
		s.repository.Filter("loan_officer_id = ?", loanOfficerID),
		s.repository.Preload("Installations"),
		s.repository.Preload("Documents"),
	}

	// Use the repository's GetAll method with the UOW and QueryProcessors
	err := s.repository.GetAll(uow, &applications, queryProcessors...)
	if err != nil {
		return nil, err
	}

	return applications, nil
}

func (s *LoanOfficerService) ProcessApplicationDecision(applicationID string, approve bool) error {
	var application loanapplication.LoanApplication
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Retrieve the loan application
	if err := s.repository.GetByID(uow, &application, applicationID); err != nil {
		return fmt.Errorf("failed to retrieve application: %w", err)
	}

	if approve {
		application.Status = "Approved"
		decisionDate := time.Now()
		application.DecisionDate = &decisionDate

		if err := s.generateInstallments(&application, uow); err != nil {
			return fmt.Errorf("failed to generate installments: %w", err)
		}
	} else {
		application.Status = "Rejected"
	}

	// Update application status
	if err := s.repository.Update(uow, &application); err != nil {
		return fmt.Errorf("failed to update application status: %w", err)
	}

	// Commit transaction only after all operations are successful
	uow.Commit()
	return nil
}

// generateInstallments creates monthly installments based on loan tenure and amount
func (s *LoanOfficerService) generateInstallments(application *loanapplication.LoanApplication, uow *repository.UOW) error {
	var loanScheme loanscheme.LoanScheme

	// Fetch the loan scheme using GetByID
	if err := s.repository.GetByID(uow, &loanScheme, application.LoanSchemeID); err != nil {
		return fmt.Errorf("failed to fetch loan scheme: %w", err)
	}
	if loanScheme.ID == 0 {
		return fmt.Errorf("loan scheme not found for ID: %d", application.LoanSchemeID)
	}

	// Calculate monthly rate and EMI
	monthlyRate := loanScheme.InterestRate / 12 / 100
	emi := (application.Amount * monthlyRate) / (1 - math.Pow(1+monthlyRate, -float64(loanScheme.Tenure)))

	// Prepare installments based on the tenure
	installments := make([]*installation.Installation, loanScheme.Tenure)
	for i := 0; i < loanScheme.Tenure; i++ {
		installmentDate := time.Now().AddDate(0, i+1, 0)
		dueDate := time.Date(
			installmentDate.Year(),
			installmentDate.Month(),
			installmentDate.Day(),
			0, 0, 0, 0,
			installmentDate.Location(),
		)

		installments[i] = &installation.Installation{
			LoanApplicationID: application.ID,
			AmountToBePaid:    emi,
			DueDate:           dueDate,
			Status:            "Pending",
		}

		// Add each installment as a separate transaction to prevent long transaction times
		if err := s.repository.Add(uow, installments[i]); err != nil {
			return fmt.Errorf("failed to save installment: %w", err)
		}
	}
	return nil
}
