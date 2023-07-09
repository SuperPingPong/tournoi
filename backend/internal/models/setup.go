package models

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var LOGGER = logger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags),
	logger.Config{
		LogLevel: logger.Info,
	})

func ConnectDatabase() (*gorm.DB, error) {
	dsn := "host=db user=postgres password=postgres dbname=database port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         LOGGER,
	})
	if err != nil {
		return nil, err
	}

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	for _, model := range ListModels() {
		err = db.AutoMigrate(model)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}
