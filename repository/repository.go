package repository

import (
	"loanApp/models/user"

	"github.com/jinzhu/gorm"
)

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
	Preload(association string) QueryProcessor
}

type GormRepositoryMySQL struct{} //->implements repository

func NewGormRepositoryMySQL() Repository {
	return &GormRepositoryMySQL{}
}

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

func (g *GormRepositoryMySQL) GetAll(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Find(out).Error
}

func (g *GormRepositoryMySQL) GetByID(uow *UOW, out interface{}, id interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.First(out, "id = ?", id).Error
}

func (g *GormRepositoryMySQL) Add(uow *UOW, out interface{}) error {
	return uow.DB.Create(out).Error
}

func (g *GormRepositoryMySQL) Update(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Save(out).Error
}

func (g *GormRepositoryMySQL) DeleteByID(uow *UOW, out interface{}, id interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Where("id = ?", id).Delete(out).Error
}

func (g *GormRepositoryMySQL) Limit(limit interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Limit(limit)
		return db, nil
	}
}

func (g *GormRepositoryMySQL) Offset(offset interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Offset(offset)
		return db, nil
	}
}

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

type UOW struct {
	DB       *gorm.DB
	Commited bool
}

func NewUnitOfWork(DB *gorm.DB) *UOW {
	return &UOW{DB: DB.Begin(), Commited: false}
}

func (u *UOW) RollBack() {
	if u.Commited {
		return
	}
	u.DB.Rollback()
}

func (u *UOW) Commit() {
	if u.Commited {
		return
	}
	u.DB.Commit()
	u.Commited = true
}

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
