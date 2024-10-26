package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"loanApp/components/customer/service"
	"loanApp/models/user"
	"loanApp/utils/log"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type RegisterController struct {
	CustomerService *service.CustomerService
	log             log.Logger
}

func NewRegisterController(customerService *service.CustomerService, log log.Logger) *RegisterController {
	return &RegisterController{
		CustomerService: customerService,
		log:             log,
	}
}

func (rc *RegisterController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/register", rc.RegisterCustomer).Methods(http.MethodPost)
}

func (rc *RegisterController) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	rc.log.Info("RegisterCustomer called")

	var newCustomer user.Customer
	if err := json.NewDecoder(r.Body).Decode(&newCustomer); err != nil {
		rc.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}
	newCustomer.Role = "Customer"

	if err := validateCustomer(newCustomer); err != nil {
		rc.log.Error("Validation error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := rc.CustomerService.CreateCustomer(&newCustomer); err != nil {
		rc.log.Error("Error creating customer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create customer")
		return
	}

	web.RespondWithJSON(w, http.StatusCreated, newCustomer)
}

func validateCustomer(customer user.Customer) error {
	if customer.Name == "" || customer.Email == "" || customer.Password == "" {
		return errors.New("name, email, and password are required")
	}
	return nil
}