package repository

import (
	"loanApp/models/user"

	"github.com/jinzhu/gorm"
)

// Repository interface with all required methods
type Repository interface {
	GetAll(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error
	GetByID(uow *UOW, out interface{}, id interface{}, queryProcessors ...QueryProcessor) error
	Add(uow *UOW, out interface{}) error
	Update(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error
	DeleteByID(uow *UOW, out interface{}, id interface{}, queryProcessors ...QueryProcessor) error
	Limit(limit interface{}) QueryProcessor
	Offset(offset interface{}) QueryProcessor
	Filter(condition string, args ...interface{}) QueryProcessor
	Count(limit, offset int, totalCount *int) QueryProcessor
	GetUserWithLoginInfo(uow *UOW, userID uint) (*user.User, error)
}

// GormRepositoryMySQL is an implementation of the Repository interface
type GormRepositoryMySQL struct{}

// NewGormRepositoryMySQL creates a new instance of GormRepositoryMySQL
func NewGormRepositoryMySQL() Repository {
	return &GormRepositoryMySQL{}
}

// executeQueryProcessors applies query processors to the GORM DB object
func executeQueryProcessors(db *gorm.DB, out interface{}, queryProcessors ...QueryProcessor) (*gorm.DB, error) {
	var err error
	for i := 0; i < len(queryProcessors); i++ {
		db, err = queryProcessors[i](db, out)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}

// Count applies a count query to get the number of records
func (g *GormRepositoryMySQL) Count(limit, offset int, totalCount *int) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		if totalCount != nil {
			err := db.Model(out).Count(totalCount).Error
			if err != nil {
				return db, err
			}
		}
		return db, nil
	}
}

// GetAll retrieves all records with optional query processors
func (g *GormRepositoryMySQL) GetAll(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Find(out).Error
}

// GetByID retrieves a record by its ID with optional query processors
func (g *GormRepositoryMySQL) GetByID(uow *UOW, out interface{}, id interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.First(out, "id = ?", id).Error
}

// Add inserts a new record into the database
func (g *GormRepositoryMySQL) Add(uow *UOW, out interface{}) error {
	return uow.DB.Create(out).Error
}

// Update modifies an existing record in the database with query processors
func (g *GormRepositoryMySQL) Update(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Save(out).Error
}

// DeleteByID removes a record from the database by its ID with optional query processors
func (g *GormRepositoryMySQL) DeleteByID(uow *UOW, out interface{}, id interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Where("id = ?", id).Delete(out).Error
}

// Limit adds a limit to the query
func (g *GormRepositoryMySQL) Limit(limit interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Limit(limit)
		return db, nil
	}
}

// Offset adds an offset to the query
func (g *GormRepositoryMySQL) Offset(offset interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Offset(offset)
		return db, nil
	}
}

// Filter adds a filter (WHERE clause) to the query
func (g *GormRepositoryMySQL) Filter(condition string, args ...interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Where(condition, args...)
		return db, nil
	}
}

func (g *GormRepositoryMySQL) Preload(association string) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Preload(association)
		return db, nil
	}
}

// UOW (Unit of Work) struct for managing database transactions
type UOW struct {
	DB       *gorm.DB
	Commited bool
}

// NewUnitOfWork starts a new transaction (Unit of Work)
func NewUnitOfWork(DB *gorm.DB) *UOW {
	return &UOW{DB: DB.Begin(), Commited: false}
}

// RollBack rolls back the transaction
func (u *UOW) RollBack() {
	if u.Commited {
		return
	}
	u.DB.Rollback()
}

// Commit commits the transaction
func (u *UOW) Commit() {
	if u.Commited {
		return
	}
	u.DB.Commit()
	u.Commited = true
}

// GetUserWithLoginInfo retrieves a user by ID and preloads the LoginInfo association
func (g *GormRepositoryMySQL) GetUserWithLoginInfo(uow *UOW, userID uint) (*user.User, error) {
	var foundUser user.User
	db, err := executeQueryProcessors(uow.DB, &foundUser, g.Preload("LoginInfo"))
	if err != nil {
		return nil, err
	}
	err = db.First(&foundUser, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &foundUser, nil
}
