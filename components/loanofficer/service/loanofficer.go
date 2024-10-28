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
	"gopkg.in/gomail.v2"
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

	// Check if the officer has any pending loan applications
	var applications []loanapplication.LoanApplication
	err := s.repository.GetAll(uow, &applications,
		s.repository.Filter("loan_officer_id = ?", officer.ID),
		s.repository.Filter("status IN (?, ?, ?)", "Pending", "PendingCollateral", "Collateral Uploaded"))
	if err != nil {
		return fmt.Errorf("failed to fetch loan applications: %w", err)
	}

	// If there are any pending applications, do not allow deletion
	if len(applications) > 0 {
		return fmt.Errorf("cannot delete loan officer with ID %s: there are pending applications", id)
	}

	// Proceed with deletion if no pending applications
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

// GetLeastLoadedOfficer returns the loan officer with the least workload based on pending loan applications
func (s *LoanOfficerService) GetLeastLoadedOfficer() (*user.LoanOfficer, error) {
	var officer user.LoanOfficer
	err := s.DB.Model(&user.LoanOfficer{}).
		Joins("LEFT JOIN loan_applications ON loan_officers.id = loan_applications.loan_officer_id AND loan_applications.status IN (?, ?, ?)", "Pending", "PendingCollateral", "Collateral Uploaded").
		Where("loan_officers.is_active = ?", true).
		Group("loan_officers.id").
		Order("COUNT(loan_applications.id) ASC").
		First(&officer).Error

	if err != nil {
		return nil, err
	}

	// Check if an officer was found
	if officer.ID == 0 {
		return nil, fmt.Errorf("no active loan officer found")
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
func (s *LoanOfficerService) ApproveInitialApplication(applicationID string, loanOfficerID uint, approve bool) error {
	var application loanapplication.LoanApplication
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Retrieve the loan application
	if err := s.repository.GetByID(uow, &application, applicationID); err != nil {
		return fmt.Errorf("failed to retrieve application: %w", err)
	}

	// Check if application is already approved, rejected, or awaiting collateral
	if application.Status == "Approved" || application.Status == "Rejected" || application.Status == "PendingCollateral" {
		return errors.New("invalid operation: application has already been processed")
	}

	if approve {
		nowtime := time.Now()
		application.DecisionDate = &nowtime      // Move to collateral document submission step
		application.Status = "PendingCollateral" // Move to collateral document submission step

		// Send email to user about the application approval
		if err := s.sendApprovalEmail(&application); err != nil {
			return fmt.Errorf("failed to send approval email: %w", err)
		}
	} else {
		application.Status = "Rejected"
	}

	// Update application status
	if err := s.repository.Update(uow, &application); err != nil {
		return fmt.Errorf("failed to update application status: %w", err)
	}

	uow.Commit()
	return nil
}

// sendReminderEmail composes and sends a reminder email using gomail and Gmail SMTP
func (s *LoanOfficerService) sendApprovalEmail(application *loanapplication.LoanApplication) error {
	// Fetch customer details associated with the installment
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	var customer user.Customer
	fmt.Println("send approval email")

	if err := s.repository.GetByID(uow, &customer, application.CustomerID); err != nil {
		return fmt.Errorf("failed to fetch customer: %w", err)
	}

	// Email content
	subject := "Loan Approval, Pending Collateral"
	message := fmt.Sprintf(`
        "Dear " + %s + ",\n\n" +
		"Your loan application (ID: %d) has been approved. Please upload your collateral documents within one week. " +
		"If the documents are not uploaded within the stipulated time, your application will be rejected.\n\n" +
		"Thank you for your attention.\n\n" +
		"Best regards,\n" +
		"Your Loan Management Team"


        Best regards,
        Loanleloplis
    `, customer.Name, application.ID)

	// Configure email settings using gomail
	m := gomail.NewMessage()
	m.SetHeader("From", "kierarieger2@gmail.com")
	m.SetHeader("To", customer.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	// Set up the Gmail SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "kierarieger2@gmail.com", "rttw twcm ponf rbtd") //<---------mera fake email and password lol

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// CheckPendingCollateralApplications checks for applications that have not uploaded collateral documents within a week
func (s *LoanOfficerService) CheckPendingCollateralApplications() error {
	var applications []loanapplication.LoanApplication
	err := s.repository.GetAll(nil, &applications,
		s.repository.Filter("status = ?", "PendingCollateral"))
	if err != nil {
		return fmt.Errorf("failed to fetch pending collateral applications: %w", err)
	}

	now := time.Now()
	for _, application := range applications {
		// Check if the decision date is set and 7 days have passed since that date
		if application.DecisionDate != nil && application.DecisionDate.Add(7*24*time.Hour).Before(now) {
			// Reject the application as the collateral documents were not uploaded
			application.Status = "Rejected"
			if err := s.repository.Update(nil, &application); err != nil {
				return fmt.Errorf("failed to reject application due to missing collateral: %w", err)
			}
			// Optionally, you may want to send a notification to the user about the rejection
			if err := s.sendRejectionEmail(&application); err != nil {
				return fmt.Errorf("failed to send rejection email: %w", err)
			}
		}
	}

	return nil
}

// sendReminderEmail composes and sends a reminder email using gomail and Gmail SMTP
func (s *LoanOfficerService) sendRejectionEmail(application *loanapplication.LoanApplication) error {
	// Fetch customer details associated with the installment
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	var customer user.Customer
	fmt.Println("send rejection email")

	if err := s.repository.GetByID(uow, &customer, application.CustomerID); err != nil {
		return fmt.Errorf("failed to fetch customer: %w", err)
	}

	// Email content
	subject := "Loan Rejection due to Pending Collateral"
	message := fmt.Sprintf(`
        "Dear " + %s + ",\n\n" +
		"Your loan application (ID: %d) has been rejected since you did not upload collateral documents within one week. " +
		"You may apply for another loan if you are still interested.\n\n" +
		"Thank you for your attention.\n\n" +
		"Best regards,\n" +
		"Your Loan Management Team"


        Best regards,
        Loanleloplis
    `, customer.Name, application.ID)

	// Configure email settings using gomail
	m := gomail.NewMessage()
	m.SetHeader("From", "kierarieger2@gmail.com")
	m.SetHeader("To", customer.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	// Set up the Gmail SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "kierarieger2@gmail.com", "rttw twcm ponf rbtd") //<---------mera fake email and password lol

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func SchedulePendingCollateralCheck(service *LoanOfficerService) {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	go func() {
		for range ticker.C {
			if err := service.CheckPendingCollateralApplications(); err != nil {
				service.log.Error("Error checking pending collateral applications: %v", err)
			}
		}
	}()
}

// ApproveCollateralDocuments approves or rejects the collateral documents for the application
func (s *LoanOfficerService) ApproveCollateralDocuments(applicationID string, loanOfficerID uint, approve bool) error {
	var application loanapplication.LoanApplication
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Retrieve the loan application
	if err := s.repository.GetByID(uow, &application, applicationID); err != nil {
		return fmt.Errorf("failed to retrieve application: %w", err)
	}

	// Check if application is in 'PendingCollateral' status before allowing collateral approval
	if application.Status != "PendingCollateral" {
		return errors.New("application is not pending collateral approval")
	}

	if approve {
		application.Status = "Approved" // Final approval of the loan
		decisionDate := time.Now()
		application.DecisionDate = &decisionDate

		// Generate installments once the loan is fully approved
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
	if err := s.sendFinalApprovalEmail(&application); err != nil {
		return fmt.Errorf("failed to send final approval email: %w", err)
	}

	uow.Commit()
	return nil
}

// sendReminderEmail composes and sends a reminder email using gomail and Gmail SMTP
func (s *LoanOfficerService) sendFinalApprovalEmail(application *loanapplication.LoanApplication) error {
	// Fetch customer details associated with the installment
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	var customer user.Customer
	fmt.Println("send final approval email")

	if err := s.repository.GetByID(uow, &customer, application.CustomerID); err != nil {
		return fmt.Errorf("failed to fetch customer: %w", err)
	}

	// Email content
	subject := "Loan Approved!"
	message := fmt.Sprintf(`
        "Dear " + %s + ",\n\n" +
		"Congratulations! Your loan application (ID: %d) has been approved. " +
		"You can login to check your installment schedule and applications." +
		"Thank you for your attention.\n\n" +
		"Best regards,\n" +
		"Your Loan Management Team"


        Best regards,
        Loanleloplis
    `, customer.Name, application.ID)

	// Configure email settings using gomail
	m := gomail.NewMessage()
	m.SetHeader("From", "kierarieger2@gmail.com")
	m.SetHeader("To", customer.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	// Set up the Gmail SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "kierarieger2@gmail.com", "rttw twcm ponf rbtd") //<---------mera fake email and password lol

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

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

	// // Prepare installments based on the tenure
	// installments := make([]*installation.Installation, loanScheme.Tenure)
	// for i := 0; i < loanScheme.Tenure; i++ {
	// 	installmentDate := time.Now().AddDate(0, i+1, 0)
	// 	dueDate := time.Date(
	// 		installmentDate.Year(),
	// 		installmentDate.Month(),
	// 		installmentDate.Day(),
	// 		0, 0, 0, 0,
	// 		installmentDate.Location(),
	// 	)

	// 	installments[i] = &installation.Installation{
	// 		LoanApplicationID: application.ID,
	// 		AmountToBePaid:    emi,
	// 		DueDate:           dueDate,
	// 		Status:            "Pending",
	// 	}

	// 	// Add each installment as a separate transaction to prevent long transaction times
	// 	if err := s.repository.Add(uow, installments[i]); err != nil {
	// 		return fmt.Errorf("failed to save installment: %w", err)
	// 	}
	// }
	// Prepare installments based on the tenure with 1-minute interval due dates
	installments := make([]*installation.Installation, loanScheme.Tenure)
	for i := 0; i < loanScheme.Tenure; i++ {
		// Set the due date to 1-minute intervals from the current time
		dueDate := time.Now().Add(time.Duration(i+1) * time.Minute)

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
func (s *LoanOfficerService) IsApplicationAssignedToOfficer(applicationID string, loanOfficerID uint) (bool, error) {
	var application loanapplication.LoanApplication
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Retrieve the loan application
	if err := s.repository.GetByID(uow, &application, applicationID); err != nil {
		return false, fmt.Errorf("failed to retrieve application: %w", err)
	}

	// Check if the LoanOfficerID matches the given officer ID
	if application.LoanOfficerID == loanOfficerID {
		return true, nil
	}

	return false, nil
}
