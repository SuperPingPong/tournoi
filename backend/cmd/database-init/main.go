package main

import (
	"fmt"

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

	if db.Exec("CREATE DATABASE database").Error != nil {
		fmt.Println("skipping database creation...")
	}

	db, err = models.ConnectDatabase()
	if err != nil {
		panic(err)
	}

}
