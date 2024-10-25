package app

import (
	"context"
	"fmt"
	"loanApp/models/loanscheme"
	"loanApp/models/logininfo"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var AllAdmins []*user.Admin
var AllLoanOfficers []*user.LoanOfficer
var AllCustomers []*user.Customer
var AllLoanSchemes []*loanscheme.LoanScheme

// Define the roles as constants
const (
	Admin       user.Role = "Admin"
	LoanOfficer user.Role = "Loan Officer"
	Customer    user.Role = "Customer"
)

type App struct {
	sync.Mutex
	Name       string
	Router     *mux.Router
	DB         *gorm.DB
	Log        log.Logger
	Server     *http.Server
	WG         *sync.WaitGroup
	Repository repository.Repository
}

func NewApp(name string, db *gorm.DB, log log.Logger,
	wg *sync.WaitGroup, repo repository.Repository) *App {
	return &App{
		Name:       name,
		DB:         db,
		Log:        log,
		WG:         wg,
		Repository: repo,
	}
}
func NewDBConnection(log log.Logger) *gorm.DB {
	db, err := gorm.Open("mysql", "root:{password}@/LoanManagementSystem?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	db.LogMode(true)
	return db
}
func (app *App) Init() {
	app.initializeRouter()
	app.initializeServer()
}
func (app *App) initializeRouter() {
	app.Log.Info(app.Name + " App Route initializing")
	app.Router = mux.NewRouter().StrictSlash(true)
	app.Router = app.Router.PathPrefix("/api/v1/loan-app").Subrouter()
}
func (app *App) initializeServer() {
	headers := handlers.AllowedHeaders([]string{
		"Content-Type", "X-Total-Count", "token",
	})
	methods := handlers.AllowedMethods([]string{
		http.MethodPost, http.MethodPut, http.MethodGet, http.MethodDelete, http.MethodOptions,
	})
	originOption := handlers.AllowedOriginValidator(app.checkOrigin)
	app.Server = &http.Server{
		Addr:         "0.0.0.0:4000",
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      handlers.CORS(headers, methods, originOption)(app.Router),
	}
	app.Log.Info("Server Exposed On 4000")
}
func (app *App) checkOrigin(origin string) bool {
	return true
}

type Controller interface {
	RegisterRoutes(router *mux.Router)
}

func (app *App) RegisterAllControllerRoutes(controllers []Controller) {
	for i := 0; i < len(controllers); i++ {
		//router are created here
		controllers[i].RegisterRoutes(app.Router)
	}
}
func (a *App) StartServer() error {
	// Initialize the server with a custom address and handler
	a.Server = &http.Server{
		Addr:    ":4000",
		Handler: a.Router, // Assume you have a router set up
	}

	fmt.Println("Server Exposed On 4000")
	return a.Server.ListenAndServe()
}

func (a *App) StopServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shut down the server gracefully
	if err := a.Server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server shut down gracefully")
}

// ClearDatabase drops specified tables in the database.
func ClearDatabase() {
	// Connect to the database
	db, err := gorm.Open("mysql", "root:{password}@tcp(localhost:3306)/LoanManagementSystem?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.GetLogger().Error("failed to connect to database: %v", err)
		return
	}
	defer db.Close() // Close the database connection when done

	// Drop the tables
	if err := db.DropTableIfExists(&user.User{}, &user.Admin{}, &user.Customer{}, &user.LoanOfficer{}, &logininfo.LoginInfo{}).Error; err != nil {
		log.GetLogger().Error("failed to drop tables: %v", err)
		return
	}

	log.GetLogger().Info("All tables dropped successfully!")
}

type ModuleConfig interface {
	TableMigration()
}

func (app *App) TableMigration(moduleConfigs []ModuleConfig) {
	for i := 0; i < len(moduleConfigs); i++ {
		moduleConfigs[i].TableMigration()
	}
}
