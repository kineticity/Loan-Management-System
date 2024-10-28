package main

import (
	"fmt"
	"loanApp/app"
	"loanApp/components/admin/service"

	loanappservice "loanApp/components/loanapplication/service"
	loanofficerService "loanApp/components/loanofficer/service"
	"loanApp/models/logininfo"
	"loanApp/models/user"
	"loanApp/modules"
	"loanApp/repository"
	"loanApp/utils/log"
	"sync"
)

func main() {
	// Logger
	logger := log.GetLogger()

	// DB Connections
	db := app.NewDBConnection(logger)
	if db == nil {
		logger.Error("Db connection failed.")
		return // Exit if DB connection fails
	}

	var wg sync.WaitGroup // Initialize WaitGroup

	repository := repository.NewGormRepositoryMySQL()
	adminService := service.NewAdminService(db, repository, logger)
	application := app.NewApp("Loan-Management-System", db, logger, &wg, repository)

	// app.ClearDatabase()

	// Initialize router and server
	application.Init()

	// Register routes
	modules.RegisterAllRoutes(application)

	// Models, migrations, etc.
	modules.ConfigureAppTables(application)

	// Create Super Admin before starting the server
	createSuperAdmin(adminService)

	// Create a new LoanApplicationService instance
	loanOfficerService := loanofficerService.NewLoanOfficerService(db, repository, logger)
	loanApplicationService := loanappservice.NewLoanApplicationService(db, repository, logger, loanOfficerService)

	// Start the NPA status check scheduler
	loanappservice.ScheduleNPAStatusCheck(loanApplicationService)
	loanappservice.StartReminderScheduler(loanApplicationService)
	loanofficerService.SchedulePendingCollateralCheck(loanOfficerService)

	// Start server in a goroutine
	wg.Add(1) // Increment the WaitGroup counter
	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes
		err := application.StartServer()
		if err != nil {
			fmt.Println(err)
			stopApp(application)
		}
	}()

	// Wait for all goroutines to finish (in this case, just the server)
	wg.Wait() // Wait here until the server stops
	stopApp(application)
}

func stopApp(app *app.App) {
	app.StopServer()
	fmt.Println("Application stopped")
}

func createSuperAdmin(adminService *service.AdminService) {
	superadmin := &user.Admin{
		User: user.User{
			Name:      "Super Admin",
			Email:     "superadmin@gmail.com",
			Password:  "password",
			IsActive:  true,
			Role:      "Admin",
			LoginInfo: []*logininfo.LoginInfo{},
		},
		LoanOfficers: []*user.LoanOfficer{},
	}

	err := adminService.CreateAdmin(superadmin)
	if err != nil {
		fmt.Println("Error creating super admin:", err)
		return
	}

	fmt.Println("Super Admin created successfully")
}
