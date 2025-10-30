package userexchanges

import (
	"errors"
	"sync"

	"vsC1Y2025V01/src/db"
	"vsC1Y2025V01/src/model"

	"gorm.io/gorm"
)

var (
	ErrExchangeNotFound     = errors.New("exchange not found")
	ErrUserExchangeNotFound = errors.New("user exchange not found")
)

type UserExchangeStore interface {
	CreateExchange(exchange *model.Exchange) error
	GetExchangeByID(id uint) (*model.Exchange, error)
	FindUserExchange(userID, exchangeID uint) (*model.UserExchange, error)
	SaveUserExchange(ue *model.UserExchange) error
	ListFormUserExchanges(userID uint) ([]model.UserExchange, error)
	DeleteUserExchange(userID, exchangeID uint) (bool, error)
}

var (
	storeMu sync.RWMutex
	store   UserExchangeStore = &gormUserExchangeStore{}
)

func SetUserExchangeStore(s UserExchangeStore) {
	storeMu.Lock()
	defer storeMu.Unlock()

	if s == nil {
		store = &gormUserExchangeStore{}
		return
	}

	store = s
}

func getUserExchangeStore() UserExchangeStore {
	storeMu.RLock()
	current := store
	storeMu.RUnlock()

	if current != nil {
		return current
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	if store == nil {
		store = &gormUserExchangeStore{}
	}

	return store
}

type gormUserExchangeStore struct{}

func (s *gormUserExchangeStore) CreateExchange(exchange *model.Exchange) error {
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	return db.DB.Create(exchange).Error
}

func (s *gormUserExchangeStore) GetExchangeByID(id uint) (*model.Exchange, error) {
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	var exchange model.Exchange
	if err := db.DB.First(&exchange, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExchangeNotFound
		}

		return nil, err
	}

	return &exchange, nil
}

func (s *gormUserExchangeStore) FindUserExchange(userID, exchangeID uint) (*model.UserExchange, error) {
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	var userExchange model.UserExchange
	if err := db.DB.Where("user_id = ? AND exchange_id = ?", userID, exchangeID).Preload("Exchange").First(&userExchange).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserExchangeNotFound
		}

		return nil, err
	}

	return &userExchange, nil
}

func (s *gormUserExchangeStore) SaveUserExchange(ue *model.UserExchange) error {
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	return db.DB.Save(ue).Error
}

func (s *gormUserExchangeStore) ListFormUserExchanges(userID uint) ([]model.UserExchange, error) {
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	var exchanges []model.UserExchange
	if err := db.DB.Preload("Exchange").Where("user_id = ? AND show_in_forms = ?", userID, true).Find(&exchanges).Error; err != nil {
		return nil, err
	}

	return exchanges, nil
}

func (s *gormUserExchangeStore) DeleteUserExchange(userID, exchangeID uint) (bool, error) {
	if db.DB == nil {
		return false, errors.New("database connection is not initialized")
	}

	res := db.DB.Where("user_id = ? AND exchange_id = ?", userID, exchangeID).Delete(&model.UserExchange{})
	if res.Error != nil {
		return false, res.Error
	}

	return res.RowsAffected > 0, nil
}
