package service

import (
	"errors"
	"loanApp/models/loanscheme"
	"loanApp/repository"
	"loanApp/utils/log"
	"loanApp/utils/web"
	"strconv"

	"github.com/jinzhu/gorm"
)

type LoanSchemeService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewLoanSchemeService(db *gorm.DB, repository repository.Repository, log log.Logger) *LoanSchemeService {
	return &LoanSchemeService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}

func (s *LoanSchemeService) CreateLoanScheme(scheme *loanscheme.LoanScheme) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	if err := validateLoanScheme(scheme); err != nil {
		return err
	}

	err := s.repository.Add(uow, scheme)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanSchemeService) GetAllLoanSchemes(allSchemes *[]*loanscheme.LoanScheme, totalCount *int, parser web.Parser) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	limit, err := strconv.Atoi(parser.Form.Get("limit"))
	if err != nil {
		limit = 10 // default value
	}

	offset, err := strconv.Atoi(parser.Form.Get("offset"))
	if err != nil {
		offset = 0 // default value
	}

	queryProcessors := []repository.QueryProcessor{
		// s.repository.Preload("CreatedBy"),
		// s.repository.Preload("UpdatedBy"),
		s.repository.Limit(limit),
		s.repository.Offset(offset),
	}

	if err := s.repository.GetAll(uow, allSchemes, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanSchemeService) UpdateLoanScheme(id string, updatedScheme *loanscheme.LoanScheme) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var scheme loanscheme.LoanScheme
	if err := s.repository.GetByID(uow, &scheme, id); err != nil {
		return err
	}

	scheme.Name = updatedScheme.Name
	scheme.Category = updatedScheme.Category
	scheme.InterestRate = updatedScheme.InterestRate
	scheme.Tenure = updatedScheme.Tenure
	scheme.UpdatedBy = updatedScheme.UpdatedBy

	if err := s.repository.Update(uow, &scheme); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanSchemeService) DeleteLoanScheme(id string) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var scheme loanscheme.LoanScheme
	if err := s.repository.GetByID(uow, &scheme, id); err != nil {
		return err
	}

	if err := s.repository.DeleteByID(uow, &scheme, id); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func validateLoanScheme(scheme *loanscheme.LoanScheme) error {
	if scheme.Name == "" {
		return errors.New("scheme name cannot be empty")
	}
	if scheme.Category == ""|| (scheme.Category!="retail"&&scheme.Category!="corporate") {
		return errors.New("category cannot be empty. Should be corporate or retail only")
	}
	if scheme.InterestRate <= 0 {
		return errors.New("interest rate must be greater than zero")
	}
	if scheme.Tenure <= 0 {
		return errors.New("tenure must be greater than zero")
	}
	return nil
}
