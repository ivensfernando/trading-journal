package db

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"time"
	"vsC1Y2025V01/src/model"
)

var DB *gorm.DB

func InitDB(logger *logrus.Entry) {
	//dsn := os.Getenv("POSTGRES_DSN")
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGPORT"),
	)
	//db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	//if err != nil {
	//	logger.WithError(err).Fatal("Failed to connect to database")
	//}
	//
	//if err := db.AutoMigrate(&model.Alert{}, &model.User{}, &model.Trade{}); err != nil {
	//	logger.WithError(err).Fatal("Failed to migrate database")
	//}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true, // Optional: enables prepared statement caching
	})
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}

	// Connection pool tuning
	sqlDB, err := db.DB()
	if err != nil {
		logger.WithError(err).Fatal("Failed to get DB from GORM")
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	if err := db.AutoMigrate(&model.Alert{}, &model.User{}, &model.Trade{}, &model.Exchange{}, &model.PairsCoins{}); err != nil {
		logger.WithError(err).Fatal("Failed to migrate database")
	}

	DB = db
	logger.Info("Database connection initialized")
}
