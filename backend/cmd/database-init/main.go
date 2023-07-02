package main

import (
	"github.com/SuperPingPong/tournoi/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=db user=postgres password=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Exec("CREATE DATABASE database")

	db, err = models.ConnectDatabase()
	if err != nil {
		panic(err)
	}

	bandNames := []string{
		"A",
		"B",
		"C",
	}
	for _, bandName := range bandNames {
		err := db.Create(&models.Band{
			Name: bandName,
		}).Error
		if err != nil {
			panic(err)
		}
	}
}
