package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"

	"github.com/SuperPingPong/tournoi/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	dsn := fmt.Sprintf("host=db user=%s password=%s dbname=database port=5432 sslmode=disable", postgresUser, postgresPassword)
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
