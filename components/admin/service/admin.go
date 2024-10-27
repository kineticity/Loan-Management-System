package service

import (
	"loanApp/app"
	"loanApp/models/statistics"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"
	"loanApp/utils/web"
	"time"

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

func (u *AdminService) GetAllAdmins(allAdmins *[]*user.Admin, totalCount *int, parser web.Parser) error {
	uow := repository.NewUnitOfWork(u.DB)
	defer uow.RollBack()

	limit, err := strconv.Atoi(parser.Form.Get("limit"))
	if err != nil {
		limit = 12 //default
	}

	offset, err := strconv.Atoi(parser.Form.Get("offset"))
	if err != nil {
		offset = 0 //default
	}

	queryProcessors := []repository.QueryProcessor{
		// u.repository.Filter("name=?", parser.Form.Get("name")),
		u.repository.Preload("LoginInfo"),
		u.repository.Preload("LoanOfficers"), //<---------------------------------------------uncommented

		u.repository.Limit(limit),
		u.repository.Offset(offset),
	}

	if err := u.repository.GetAll(uow, allAdmins, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *AdminService) GetStatistics(startDate, endDate time.Time) (*statistics.Statistics, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Get count of active users
	activeUsersCount, err := s.repository.GetActiveUsersCount(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate average session time
	loginInfos, err := s.repository.GetLoginInfosForCustomers(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var totalSessionTime float64
	var sessionCount int
	for _, loginInfo := range loginInfos {
		if loginInfo.LogoutTime != nil {
			totalSessionTime += loginInfo.LogoutTime.Sub(loginInfo.LoginTime).Minutes()
			sessionCount++
		}
	}

	averageSessionTime := 0.0
	if sessionCount > 0 {
		averageSessionTime = totalSessionTime / float64(sessionCount)
	}

	// Get total loan applications count
	loanApplicationsCount, err := s.repository.GetTotalLoanApplicationsCount(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get NPA loan applications count
	npaCount, err := s.repository.GetNPACount(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get statistics per loan scheme
	loanSchemeStats, err := s.repository.GetLoanSchemeStats(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get statistics per loan officer
	loanOfficerStats, err := s.repository.GetLoanOfficerStats(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	stats := &statistics.Statistics{
		ActiveCustomersCount:       int(activeUsersCount),
		AverageCustomerSessionTime: averageSessionTime,
		TotalLoanApplications:      int(loanApplicationsCount),
		NPACount:                   int(npaCount),
		SchemeStats:                loanSchemeStats,
		OfficerStats:               loanOfficerStats,
	}

	uow.Commit()
	return stats, nil
}
