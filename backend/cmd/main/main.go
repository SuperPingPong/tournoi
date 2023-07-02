package main

import (
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

	public.NewAPI(db, r, &http.Client{})

	err = r.Run("0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
}
