package service

import (
	"errors"
	"fmt"
	"loanApp/components/loanofficer/service"
	"loanApp/models/document"
	"loanApp/models/installation"
	"loanApp/models/loanapplication"
	"loanApp/models/user"
	"time"

	"loanApp/repository"
	"loanApp/utils/log"

	"gopkg.in/gomail.v2"

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
const DocumentUploadDir = "C:\\Users\\keertana.kalathingal\\Documents\\LMS 3\\4 Loan Management System 26th given final\\4 Loan Management System\\uploads"

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

func (s *LoanApplicationService) PayInstallment(customerID, loanApplicationID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Find the installment with the nearest due date that is unpaid
	var installment installation.Installation
	err := s.repository.GetAll(uow, &installment,
		s.repository.Filter("loan_application_id = ? AND status = ?", loanApplicationID, "Pending"),
		s.repository.OrderBy("due_date ASC"), // Order by due date in ascending order
		s.repository.Limit(1),
	)
	if err != nil || installment.ID == 0 {
		var loanapp loanapplication.LoanApplication

		err := s.repository.GetByID(uow, &loanapp, loanApplicationID)
		if err != nil {
			return errors.New("loan application not found")
		}

		loanapp.Status = "Paid Off"
		if err := s.repository.Update(uow, &loanapp); err != nil {
			return errors.New("failed to update loan application status to paid off")
		}
		return errors.New("no pending installments found for this loan application")

	}

	// Update the installment as paid
	installment.Status = "Paid"
	now := time.Now()
	installment.PaymentDate = &now

	// Save the updated installment
	if err := s.repository.Update(uow, &installment); err != nil {
		return errors.New("failed to update installment payment status")
	}

	uow.Commit()
	return nil
}

// CheckForNPA checks loan applications for Non-Performing Asset (NPA) status
func (s *LoanApplicationService) CheckForNPA() error {
	var applications []loanapplication.LoanApplication

	// Start a new Unit of Work
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Get all active loan applications
	if err := s.repository.GetAll(uow, &applications); err != nil {
		return fmt.Errorf("failed to fetch loan applications: %w", err)
	}

	for _, app := range applications {
		// Fetch the last three due installments for each loan application
		var installments []installation.Installation
		err := s.repository.GetAll(uow, &installments,
			s.repository.Filter("loan_application_id= ?", app.ID),
			s.repository.Filter("status= ?", "Pending"),

			s.repository.OrderBy("due_date DESC"),
			s.repository.Limit(3))
		//  )
		fmt.Println("Installments:", installments)
		if err != nil {
			return fmt.Errorf("failed to fetch installments: %w", err)
		}

		fmt.Println("checking overdues for", app.ID)

		// Check if the last three installments are overdue
		overdueCount := 0
		now := time.Now()
		for _, inst := range installments {
			fmt.Println(inst.DueDate, " ", now)
			if inst.DueDate.Before(now) {
				overdueCount++
			}
		}

		fmt.Println("overdue installments= ", overdueCount)

		// Mark as NPA if the last three installments are overdue
		if overdueCount == 3 && !app.IsNPA {
			app.IsNPA = true
			if err := s.repository.Update(uow, &app); err != nil {
				return fmt.Errorf("failed to update loan application NPA status: %w", err)
			}
		}
	}

	// Commit changes
	uow.Commit()
	return nil
}

func ScheduleNPAStatusCheck(service *LoanApplicationService) {
	ticker := time.NewTicker(1 * time.Minute) // Run daily
	service.log.Info("GOROUTINE TO CHECK NPA STATUS")
	go func() {
		for range ticker.C {
			err := service.CheckForNPA()
			if err != nil {
				log.GetLogger().Error("Error running NPA check: %v", err)
			}
		}
	}()
}

// ScheduleReminders fetches installments due in 2 days and sends reminder emails
func ScheduleReminders(s *LoanApplicationService) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Define due date range for 2 days from now
	dueDate := time.Now() //due today checking rn change later to 2 days from now
	//.AddDate(0, 0, 2) // 2 days from now
	startOfDay := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var installments []installation.Installation
	err := s.repository.GetAll(uow, &installments,
		s.repository.Filter("status = ?", "Pending"),
		s.repository.Filter("due_date >= ? AND due_date < ?", startOfDay, endOfDay),
	)
	if err != nil {
		return fmt.Errorf("failed to fetch due installments: %w", err)
	}

	// Send reminder email for each installment
	for _, installment := range installments {
		if err := s.sendReminderEmail(installment); err != nil {
			fmt.Printf("failed to send reminder for installment %d: %v\n", installment.ID, err)
		}
	}

	return nil
}

// sendReminderEmail composes and sends a reminder email using gomail and Gmail SMTP
func (s *LoanApplicationService) sendReminderEmail(installment installation.Installation) error {
	// Fetch customer details associated with the installment
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	var customer user.Customer
	fmt.Println("senreminderemail")
	var loanapplication loanapplication.LoanApplication
	err := s.repository.GetByID(uow, &loanapplication, installment.LoanApplicationID)
	if err != nil {
		return fmt.Errorf("failed to fetch loan application: %w", err)
	}
	if err := s.repository.GetByID(uow, &customer, loanapplication.CustomerID); err != nil {
		return fmt.Errorf("failed to fetch customer: %w", err)
	}

	// Email content
	subject := "Loan Payment Reminder"
	message := fmt.Sprintf(`
        Dear %s,

        This is a reminder that your installment of $%.2f is due on %s.

        Loan Details:
        Loan Application ID: %d
        Amount Due: $%.2f
        Due Date: %s

        Paisa dedo guis.

        Best regards,
        Loanleloplis
    `, customer.Name, installment.AmountToBePaid, installment.DueDate.Format("2006-01-02"), installment.LoanApplicationID, installment.AmountToBePaid, installment.DueDate.Format("2006-01-02"))

	// Configure email settings using gomail
	m := gomail.NewMessage()
	m.SetHeader("From", "kierarieger2@gmail.com")
	m.SetHeader("To", customer.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	// Set up the Gmail SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "kierarieger2@gmail.com", "rttw twcm ponf rbtd")//<---------mera fake email and password lol

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// StartReminderScheduler starts a goroutine to schedule reminders at midnight and repeat every 24 hours
func StartReminderScheduler(s *LoanApplicationService) {
	// now := time.Now()
	// nextMidnight := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	// durationUntilMidnight := nextMidnight.Sub(now)

	// Start a goroutine to wait until midnight, then run every 24 hours
	go func() {
		time.Sleep(2*time.Minute) //every 2 min currently
		for {
			if err := ScheduleReminders(s); err != nil {
				log.GetLogger().Error("Error scheduling reminders: %v", err)
			}
			time.Sleep(24 * time.Hour) // Sleep for 24 hours
		}
	}()
}
