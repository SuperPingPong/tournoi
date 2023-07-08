package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/SuperPingPong/tournoi/internal/controllers/public"
	"github.com/SuperPingPong/tournoi/internal/models"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := models.ConnectDatabase()
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	public.NewAPI(db, r, &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})

	bands := []models.Band{
		{
			Name: "A",
			Day:  1,
		},
		{
			Name: "B",
			Day:  1,
		},
		{
			Name: "C",
			Day:  1,
		},
		{
			Name: "D",
			Day:  1,
		},
		{
			Name: "E",
			Day:  1,
		},
		{
			Name: "F",
			Day:  1,
		},
		{
			Name: "G",
			Day:  1,
		},
		{
			Name: "H",
			Day:  1,
		},
		{
			Name: "1",
			Day:  2,
		},
		{
			Name: "2",
			Day:  2,
		},
		{
			Name: "3",
			Day:  2,
		},
		{
			Name: "4",
			Day:  2,
		},
		{
			Name: "5",
			Day:  2,
		},
		{
			Name: "6",
			Day:  2,
		},
		{
			Name: "7",
			Day:  2,
		},
	}
	for _, band := range bands {
		err := db.Create(&band).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrDuplicatedKey) {
				panic(err)
			}
			fmt.Printf("skipping insertion of band %s because of duplicate key\n", band.Name)
		}
	}

	err = r.Run("0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
}
