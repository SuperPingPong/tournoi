package main

import (
	"crypto/tls"
	"net/http"

	"github.com/SuperPingPong/tournoi/internal/controllers/public"
	"github.com/SuperPingPong/tournoi/internal/models"

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
			Name: "1",
			Day:  2,
		},
		{
			Name: "2",
			Day:  2,
		},
	}
	for _, band := range bands {
		err := db.Create(&band).Error
		if err != nil {
			panic(err)
		}
	}

	err = r.Run("0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
}
