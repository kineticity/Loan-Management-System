package controller

import (
	"encoding/json"
	"net/http"

	"loanApp/components/loanofficer/service"
	"loanApp/components/middleware"
	"loanApp/models/user"
	"loanApp/utils/log"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type LoanOfficerController struct {
	LoanOfficerService *service.LoanOfficerService
	log                log.Logger
}

func NewLoanOfficerController(loanOfficerService *service.LoanOfficerService, log log.Logger) *LoanOfficerController {
	return &LoanOfficerController{
		LoanOfficerService: loanOfficerService,
		log:                log,
	}
}

func (c *LoanOfficerController) RegisterRoutes(router *mux.Router) {
	officerRouter := router.PathPrefix("/loan-officer").Subrouter()
	officerRouter.Use(middleware.TokenAuthMiddleware)
	officerRouter.Use(middleware.AdminOnly) // Admin authorization middleware applied globally
	officerRouter.HandleFunc("/", c.CreateLoanOfficer).Methods(http.MethodPost)
	officerRouter.HandleFunc("/", c.GetAllLoanOfficers).Methods(http.MethodGet)
	officerRouter.HandleFunc("/{id}", c.UpdateLoanOfficer).Methods(http.MethodPut)
	officerRouter.HandleFunc("/{id}", c.DeleteLoanOfficer).Methods(http.MethodDelete)
}

func (c *LoanOfficerController) CreateLoanOfficer(w http.ResponseWriter, r *http.Request) {
	var newOfficer user.LoanOfficer
	if err := json.NewDecoder(r.Body).Decode(&newOfficer); err != nil {
		c.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	newOfficer.Role = "Loan Officer"
	if err := c.LoanOfficerService.CreateLoanOfficer(&newOfficer); err != nil {
		c.log.Error("Error creating loan officer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create loan officer")
		return
	}

	web.RespondWithJSON(w, http.StatusCreated, newOfficer)

}

func (c *LoanOfficerController) GetAllLoanOfficers(w http.ResponseWriter, r *http.Request) {
	loanOfficers, err := c.LoanOfficerService.GetAllLoanOfficers()
	if err != nil {
		c.log.Error("Error fetching loan officers: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch loan officers")
		return
	}
	web.RespondWithJSON(w, http.StatusOK, loanOfficers)
}

func (c *LoanOfficerController) UpdateLoanOfficer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	officerID := vars["id"]

	var updatedOfficer user.LoanOfficer
	if err := json.NewDecoder(r.Body).Decode(&updatedOfficer); err != nil {
		c.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if err := c.LoanOfficerService.UpdateLoanOfficer(officerID, &updatedOfficer); err != nil {
		c.log.Error("Error updating loan officer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not update loan officer")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, updatedOfficer)
}

func (c *LoanOfficerController) DeleteLoanOfficer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	officerID := vars["id"]

	if err := c.LoanOfficerService.DeleteLoanOfficer(officerID); err != nil {
		c.log.Error("Error deleting loan officer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not delete loan officer")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, "Loan officer deleted successfully")
}
