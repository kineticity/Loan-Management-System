package controller

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"loanApp/components/loanapplication/service"
	"loanApp/components/middleware"
	"loanApp/models/document"
	"loanApp/models/loanapplication"
	"loanApp/utils/log"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type LoanApplicationController struct {
	LoanAppService *service.LoanApplicationService
	log            log.Logger
}

func NewLoanApplicationController(loanAppService *service.LoanApplicationService, log log.Logger) *LoanApplicationController {
	return &LoanApplicationController{
		LoanAppService: loanAppService,
		log:            log,
	}
}

func (c *LoanApplicationController) RegisterRoutes(router *mux.Router) {
	loanAppRouter := router.PathPrefix("/loan-applications").Subrouter()
	loanAppRouter.Use(middleware.TokenAuthMiddleware)
	loanAppRouter.Use(middleware.CustomerOnly)
	loanAppRouter.HandleFunc("", c.CreateLoanApplicationWithDocs).Methods(http.MethodPost)
	loanAppRouter.HandleFunc("", c.GetCustomerLoanApplications).Methods(http.MethodGet)
}

// todo: loan officer assign with minimum load , loanschemeid verify
func (c *LoanApplicationController) CreateLoanApplicationWithDocs(w http.ResponseWriter, r *http.Request) {
	c.log.Info("CreateLoanApplicationWithDocs called")

	if err := r.ParseMultipartForm(10 << 20); err != nil { // Limit to 10MB
		web.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	customerID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("Unauthorized access: ", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	loanSchemeIDStr := r.FormValue("loan_scheme_id")
	loanSchemeID64, err := strconv.ParseUint(loanSchemeIDStr, 10, 32)
	if err != nil {
		c.log.Error("Invalid loan_scheme_id: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid loan scheme ID")
		return
	}
	loanSchemeID := uint(loanSchemeID64)

	amountStr := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.log.Error("Invalid amount: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid amount")
		return
	}

	application := loanapplication.LoanApplication{
		CustomerID:      customerID,
		Status:          "Pending",
		Amount:          amount,
		LoanSchemeID:    loanSchemeID,
		ApplicationDate: time.Now(),
		IsNPA:           false,
	}

	//todo: make separate keys for documents, chanfe filename/path to include user id,loan app id
	var documents []*document.Document
	files := r.MultipartForm.File["documents"] // key=documents
	if len(files) == 0 {
		c.log.Error("No documents provided")
		web.RespondWithError(w, http.StatusBadRequest, "No documents provided")
		return
	}
	c.log.Info("Number of documents received: ", len(files))

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			c.log.Error("Error opening file: ", err)
			continue
		}
		defer file.Close()

		fileName := fileHeader.Filename
		filePath := filepath.Join(service.DocumentUploadDir, fileName)

		dst, err := os.Create(filePath) //creates a file in path
		if err != nil {
			c.log.Error("Error saving file: ", err)
			continue
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil { //copies file content to created file
			c.log.Error("Error copying file: ", err)
			continue
		}

		doc := &document.Document{
			DocumentType: fileHeader.Header.Get("Content-Type"),
			URL:          filePath,
		}
		documents = append(documents, doc)
	}

	if err := c.LoanAppService.CreateLoanApplicationWithDocs(&application, documents); err != nil {
		c.log.Error("Error creating loan application: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create loan application")
		return
	}

	web.RespondWithJSON(w, http.StatusCreated, application)
}

func (c *LoanApplicationController) GetCustomerLoanApplications(w http.ResponseWriter, r *http.Request) {
	c.log.Info("GetCustomerLoanApplications called")
	//customer only authorized to get own applications
	customerID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("Unauthorized access: ", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	var applications []loanapplication.LoanApplication
	if err := c.LoanAppService.GetLoanApplicationsByCustomer(customerID, &applications); err != nil {
		c.log.Error("Error fetching loan applications: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch loan applications")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, applications)
}
