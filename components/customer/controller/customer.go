// package controller

// import (
// 	"encoding/json"
// 	"errors"
// 	"net/http"

// 	"loanApp/app"
// 	"loanApp/components/customer/service"
// 	"loanApp/models/user"
// 	"loanApp/utils/log"
// 	"loanApp/utils/web"

// 	"github.com/gorilla/mux"
// )

// type CustomerController struct {
// 	CustomerService *service.CustomerService
// 	log             log.Logger
// }

// func NewCustomerController(CustomerService *service.CustomerService, log log.Logger) *CustomerController {
// 	return &CustomerController{
// 		CustomerService: CustomerService,
// 		log:             log,
// 	}
// }

// func (c *CustomerController) RegisterRoutes(router *mux.Router) {
// 	customerRouter := router.PathPrefix("/customer").Subrouter()
// 	customerRouter.HandleFunc("/register", c.RegisterCustomer).Methods(http.MethodPost)
// 	customerRouter.HandleFunc("/update", c.UpdateCustomer).Methods(http.MethodPut)
// }

// // RegisterCustomer allows customers to register
// func (c *CustomerController) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
// 	c.log.Info("RegisterCustomer called")

// 	var newCustomer user.Customer
// 	if err := json.NewDecoder(r.Body).Decode(&newCustomer); err != nil {
// 		c.log.Error("Invalid input: ", err)
// 		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
// 		return
// 	}
// 	newCustomer.Role = "Customer"

// 	// Validate customer data
// 	if err := validateCustomer(newCustomer); err != nil {
// 		c.log.Error("Validation error: ", err)
// 		web.RespondWithError(w, http.StatusBadRequest, err.Error())
// 		return
// 	}

// 	// Hash the password using bcrypt
// 	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newCustomer.Password), bcrypt.DefaultCost)
// 	// if err != nil {
// 	// 	c.log.Error("Error hashing password: ", err)
// 	// 	web.RespondWithError(w, http.StatusInternalServerError, "Could not hash password")
// 	// 	return
// 	// }
// 	// newCustomer.Password = string(hashedPassword)

// 	// Create customer
// 	if _, err := c.CustomerService.CreateCustomer(&newCustomer); err != nil {
// 		c.log.Error("Error creating customer: ", err)
// 		web.RespondWithError(w, http.StatusInternalServerError, "Could not create customer")
// 		return
// 	}

// 	app.AllCustomers = append(app.AllCustomers, &newCustomer)

// 	web.RespondWithJSON(w, http.StatusCreated, newCustomer)
// }

// // UpdateCustomer allows authenticated customers to update their account
// func (c *CustomerController) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
// 	c.log.Info("UpdateCustomer called")

// 	// Assuming some middleware has authenticated the user and provided user ID in the request context
// 	customerID, err := web.GetUserIDFromContext(r)
// 	if err != nil {
// 		c.log.Error("Unauthorized access: ", err)
// 		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
// 		return
// 	}

// 	var updatedCustomer user.Customer
// 	if err := json.NewDecoder(r.Body).Decode(&updatedCustomer); err != nil {
// 		c.log.Error("Invalid input: ", err)
// 		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
// 		return
// 	}

// 	// Fetch the customer from the database
// 	existingCustomer, err := c.CustomerService.GetCustomerByID(customerID)
// 	if err != nil {
// 		c.log.Error("Error fetching customer: ", err)
// 		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch customer")
// 		return
// 	}

// 	// Update fields (e.g., name, email, etc.), but password update can be handled separately
// 	if updatedCustomer.Name != "" {
// 		existingCustomer.Name = updatedCustomer.Name
// 	}
// 	if updatedCustomer.Email != "" {
// 		existingCustomer.Email = updatedCustomer.Email
// 	}

// 	// Update the customer in the database
// 	if err := c.CustomerService.UpdateCustomer(existingCustomer); err != nil {
// 		c.log.Error("Error updating customer: ", err)
// 		web.RespondWithError(w, http.StatusInternalServerError, "Could not update customer")
// 		return
// 	}

// 	web.RespondWithJSON(w, http.StatusOK, existingCustomer)
// }

// // validateCustomer validates the customer data
// func validateCustomer(customer user.Customer) error {
// 	if customer.Name == "" {
// 		return errors.New("name cannot be empty")
// 	}
// 	if customer.Email == "" {
// 		return errors.New("email cannot be empty")
// 	}
// 	if customer.Password == "" {
// 		return errors.New("password cannot be empty")
// 	}
// 	// Additional validations can be added as needed
// 	return nil
// }

package controller

import (
	"encoding/json"
	"net/http"

	"loanApp/components/customer/service"
	"loanApp/components/middleware"
	"loanApp/models/user"
	"loanApp/utils/log"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type CustomerController struct {
	CustomerService *service.CustomerService
	log             log.Logger
}

func NewCustomerController(customerService *service.CustomerService, log log.Logger) *CustomerController {
	return &CustomerController{
		CustomerService: customerService,
		log:             log,
	}
}

func (c *CustomerController) RegisterRoutes(router *mux.Router) {
	customerRouter := router.PathPrefix("/customer").Subrouter()
	customerRouter.Use(middleware.TokenAuthMiddleware)
	customerRouter.Use(middleware.CustomerOnly) // Customer Authorization
	customerRouter.HandleFunc("/update", c.UpdateCustomer).Methods(http.MethodPut)
}


//update
func (c *CustomerController) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	c.log.Info("UpdateCustomer called")

	// Get logged in customer's ID
	customerID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("Unauthorized access: ", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	var updatedCustomer user.Customer
	if err := json.NewDecoder(r.Body).Decode(&updatedCustomer); err != nil {
		c.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	existingCustomer, err := c.CustomerService.GetCustomerByID(customerID)
	if err != nil {
		c.log.Error("Error fetching customer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch customer")
		return
	}

	if updatedCustomer.Name != "" {
		existingCustomer.Name = updatedCustomer.Name
	}
	if updatedCustomer.Email != "" {
		existingCustomer.Email = updatedCustomer.Email
	}

	// Save updates
	if err := c.CustomerService.UpdateCustomer(existingCustomer); err != nil {
		c.log.Error("Error updating customer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not update customer")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, existingCustomer)
}
