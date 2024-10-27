package service

import (
	"errors"
	"loanApp/components/loanofficer/service"
	"loanApp/models/document"
	"loanApp/models/loanapplication"
	"loanApp/models/user"
	"time"

	"loanApp/repository"
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type LoanApplicationService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewLoanApplicationService(db *gorm.DB, repository repository.Repository, log log.Logger) *LoanApplicationService {
	return &LoanApplicationService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}

// local folder for documents
const DocumentUploadDir = "C:\\Users\\Dev.patel\\Downloads\\4 Loan Management System 26th given final\\4 Loan Management System\\uploads"

func (s *LoanApplicationService) CreateLoanApplicationWithDocs(application *loanapplication.LoanApplication, docs []*document.Document) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	application.ApplicationDate = time.Now()

	var officer *user.LoanOfficer
	ser := service.NewLoanOfficerService(s.DB, s.repository, s.log)

	officer, err := ser.GetLeastLoadedOfficer() //officer has least workload
	if err != nil {
		return err
	}
	application.LoanOfficerID = officer.ID

	if err := s.repository.Add(uow, application); err != nil {
		return err
	}
	officer.AssignedLoans = append(officer.AssignedLoans, application) //
	if err := s.repository.Update(uow, officer); err != nil {
		return err
	}

	for _, doc := range docs {
		doc.LoanApplicationID = application.ID
		if err := s.repository.Add(uow, doc); err != nil {
			return err
		}
	}

	uow.Commit()
	return nil
}

func (s *LoanApplicationService) GetLoanApplicationsByCustomer(customerID uint, applications *[]loanapplication.LoanApplication) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	queryProcessors := []repository.QueryProcessor{
		s.repository.Filter("customer_id = ?", customerID),
		// s.repository.Preload("Installations"), //uncomment later
		s.repository.Preload("Documents"),
		s.repository.Preload("Installations"),
	}

	if err := s.repository.GetAll(uow, applications, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanApplicationService) UpdateLoanApplicationStatus(loanApplicationID uint, loanOfficerID uint, newStatus string) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var application loanapplication.LoanApplication
	if err := s.repository.GetByID(uow, &application, loanApplicationID); err != nil {
		return err
	}

	if application.LoanOfficerID != loanOfficerID {
		return errors.New("loan officer not authorized to update this application")
	}

	if application.Status != newStatus {
		application.Status = newStatus
		*application.DecisionDate = time.Now()
	}

	if err := s.repository.Update(uow, &application); err != nil {
		return err
	}

	uow.Commit()
	return nil
}
