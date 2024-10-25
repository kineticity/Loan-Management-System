package service

import (
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type CustomerService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewCustomerService(db *gorm.DB, repo repository.Repository, logger log.Logger) *CustomerService {
	return &CustomerService{
		DB:         db,
		repository: repo,
		log:        logger,
	}
}

// CreateCustomer creates a new customer in the database
func (s *CustomerService) CreateCustomer(customer *user.Customer) (*user.Customer, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	err := s.repository.Add(uow, &customer.User) // Add user first
	if err != nil {
		return nil, err
	}

	// Then, use the same User ID to create the Admin
	customer.ID = customer.User.ID        // Link admin to the user by setting UserID
	err = s.repository.Add(uow, customer) // Add admin
	if err != nil {
		return nil, err
	}

	uow.Commit()
	return customer, nil
}

// UpdateCustomer updates customer details in the database
func (s *CustomerService) UpdateCustomer(customer *user.Customer) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Update customer record
	if err := s.repository.Update(uow, customer); err != nil {
		s.log.Error("Error updating customer: ", err)
		return err
	}

	uow.Commit()
	return nil
}

// GetCustomerByID retrieves customer data by customer ID
func (s *CustomerService) GetCustomerByID(customerID uint) (*user.Customer, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var customer user.Customer
	if err := s.repository.GetByID(uow, &customer, customerID); err != nil {
		s.repository.Preload("LoginInfo")
		s.log.Error("Error fetching customer: ", err)
		return nil, err
	}

	return &customer, nil
}

// // DeleteCustomer removes a customer from the database
// func (s *CustomerService) DeleteCustomer(customerID uint) error {
// 	uow := repository.NewUnitOfWork(s.DB)
// 	defer uow.RollBack()

// 	// Find the customer
// 	var customer user.Customer
// 	if err := s.repository.GetByID(uow, &customer, customerID); err != nil {
// 		s.log.Error("Error fetching customer: ", err)
// 		return err
// 	}

// 	// Delete the customer record
// 	if err := s.repository.Delete(uow, &customer); err != nil {
// 		s.log.Error("Error deleting customer: ", err)
// 		return err
// 	}

// 	uow.Commit()
// 	return nil
// }
