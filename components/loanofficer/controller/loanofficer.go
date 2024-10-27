package controller

import (
	"encoding/json"
	"net/http"

	"loanApp/app"
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
	officerRouter.Use(middleware.AdminOnly)

	officerRouter.HandleFunc("/", c.CreateLoanOfficer).Methods(http.MethodPost)
	officerRouter.HandleFunc("/", c.GetAllLoanOfficers).Methods(http.MethodGet)
	officerRouter.HandleFunc("/{id}", c.UpdateLoanOfficer).Methods(http.MethodPut)
	officerRouter.HandleFunc("/{id}", c.DeleteLoanOfficer).Methods(http.MethodDelete)

	decisionRouter := router.PathPrefix("/loan-officer/decision").Subrouter()
	decisionRouter.Use(middleware.TokenAuthMiddleware)
	decisionRouter.Use(middleware.LoanOfficerOnly)
	decisionRouter.HandleFunc("/applications", c.GetAssignedLoanApplications).Methods(http.MethodGet)
	decisionRouter.HandleFunc("/applications/{id}", c.ApproveOrRejectApplication).Methods(http.MethodPost)
}

func (c *LoanOfficerController) CreateLoanOfficer(w http.ResponseWriter, r *http.Request) {
	var newOfficer user.LoanOfficer
	if err := json.NewDecoder(r.Body).Decode(&newOfficer); err != nil {
		c.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}
	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("No such admin found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No admin found")
		return
	}
	var admin *user.Admin
	for _, a := range app.AllAdmins {
		if a.ID == userID {
			admin = a
		}
	}

	if admin == nil {
		c.log.Error("No such admin found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No admin found")
		return
	}

	newOfficer.CreatedByAdminID = admin.ID
	newOfficer.Role = "Loan Officer"
	if err := c.LoanOfficerService.CreateLoanOfficer(&newOfficer); err != nil {
		c.log.Error("Error creating loan officer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create loan officer")
		return
	}

	web.RespondWithJSON(w, http.StatusCreated, newOfficer)

}

func (c *LoanOfficerController) GetAllLoanOfficers(w http.ResponseWriter, r *http.Request) {
	c.log.Info("GetAllLoanOfficers called")

	parser := web.NewParser(r)
	allLoanOfficers := []*user.LoanOfficer{}
	var totalCount int

	if err := c.LoanOfficerService.GetAllLoanOfficers(&allLoanOfficers, &totalCount, *parser); err != nil {
		c.log.Error("Error fetching officers: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch officers")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, allLoanOfficers)
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

	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("No such admin found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No admin found")
		return
	}
	var admin *user.Admin
	for _, a := range app.AllAdmins {
		if a.ID == userID {
			admin = a
		}
	}
	if len(updatedOfficer.UpdatedBy) == 0 {
		updatedOfficer.UpdatedBy = []*user.Admin{admin}
	} else {
		updatedOfficer.UpdatedBy = append(updatedOfficer.UpdatedBy, admin)
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

func (c *LoanOfficerController) GetAssignedLoanApplications(w http.ResponseWriter, r *http.Request) {
	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("User ID not found in context:", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	applications, err := c.LoanOfficerService.GetAssignedLoanApplications(userID)
	if err != nil {
		c.log.Error("Error fetching loan applications:", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch loan applications")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, applications)
}
func (c *LoanOfficerController) ApproveOrRejectApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["id"]

	// Get the user ID of the loan officer from the request context
	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("User ID not found in context:", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	// Check if the application is assigned to the loan officer
	isAssigned, err := c.LoanOfficerService.IsApplicationAssignedToOfficer(applicationID, userID)
	if err != nil {
		c.log.Error("Error checking application assignment:", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not verify application assignment")
		return
	}
	if !isAssigned {
		c.log.Error("Unauthorized access: Application not assigned to this officer")
		web.RespondWithError(w, http.StatusForbidden, "You are not authorized to make this decision")
		return
	}

	var decision struct {
		Approve bool `json:"approve"`
	}
	if err := json.NewDecoder(r.Body).Decode(&decision); err != nil {
		c.log.Error("Invalid input:", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	err = c.LoanOfficerService.ProcessApplicationDecision(applicationID, decision.Approve)
	if err != nil {
		c.log.Error("Error processing loan application decision:", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Error processing application decision")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, "Application processed successfully")
}
