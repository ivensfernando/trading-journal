package auth

import (
	"errors"
	"sync"
	"time"

	"vsC1Y2025V01/src/db"
	"vsC1Y2025V01/src/model"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
	Update(user *model.User) error
}

var (
	userRepo   UserRepository = &gormUserRepository{}
	userRepoMu sync.RWMutex
)

func SetUserRepository(repo UserRepository) {
	userRepoMu.Lock()
	defer userRepoMu.Unlock()

	if repo == nil {
		userRepo = &gormUserRepository{}
		return
	}

	userRepo = repo
}

func getUserRepository() UserRepository {
	userRepoMu.RLock()
	repo := userRepo
	userRepoMu.RUnlock()

	if repo != nil {
		return repo
	}

	userRepoMu.Lock()
	defer userRepoMu.Unlock()

	if userRepo == nil {
		userRepo = &gormUserRepository{}
	}

	return userRepo
}

type gormUserRepository struct{}

func (r *gormUserRepository) Create(user *model.User) error {
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	return db.DB.Create(user).Error
}

func (r *gormUserRepository) FindByUsername(username string) (*model.User, error) {
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	var user model.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *gormUserRepository) FindByID(id uint) (*model.User, error) {
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *gormUserRepository) Update(user *model.User) error {
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	if user != nil {
		user.UpdatedAt = time.Now()
	}

	return db.DB.Save(user).Error
}
