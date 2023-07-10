package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/SuperPingPong/tournoi/internal/controllers/public"
	"github.com/SuperPingPong/tournoi/internal/models"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	db, err := models.ConnectDatabase()
	if err != nil {
		panic(err)
	}

	_ = godotenv.Load()
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		log.Fatal("ADMIN_EMAIL environment variable is empty")
	}

	r := gin.Default()

	public.NewAPI(db, r, &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})

	bands := []models.Band{
		{
			Name:       "A",
			Day:        1,
			Color:      "blue",
			Sex:        "ALL",
			MaxPoints:  699,
			MaxEntries: 72,
			Price:      9,
		},
		{
			Name:       "B",
			Day:        1,
			Color:      "blue",
			Sex:        "ALL",
			MaxPoints:  1199,
			MaxEntries: 72,
			Price:      9,
		},
		{
			Name:       "C",
			Day:        1,
			Color:      "pink",
			Sex:        "ALL",
			MaxPoints:  799,
			MaxEntries: 72,
			Price:      9,
		},
		{
			Name:       "D",
			Day:        1,
			Color:      "pink",
			Sex:        "ALL",
			MaxPoints:  1399,
			MaxEntries: 72,
			Price:      9,
		},
		{
			Name:       "E",
			Day:        1,
			Color:      "yellow",
			Sex:        "F",
			MaxPoints:  1199,
			MaxEntries: 72,
			Price:      9,
		},
		{
			Name:       "F",
			Day:        1,
			Color:      "green",
			Sex:        "ALL",
			MaxPoints:  999,
			MaxEntries: 72,
			Price:      9,
		},
		{
			Name:       "G",
			Day:        1,
			Color:      "green",
			Sex:        "ALL",
			MaxPoints:  1599,
			MaxEntries: 72,
			Price:      9,
		},
		{
			Name:       "1",
			Day:        2,
			Color:      "blue",
			Sex:        "ALL",
			MaxPoints:  1099,
			MaxEntries: 72,
			Price:      9,
		},
		{
			Name:       "2",
			Day:        2,
			Color:      "blue",
			Sex:        "ALL",
			MaxPoints:  1699,
			MaxEntries: 72,
			Price:      10,
		},
		{
			Name:       "3",
			Day:        2,
			Color:      "pink",
			Sex:        "ALL",
			MaxPoints:  1299,
			MaxEntries: 72,
			Price:      9,
		},
		{
			Name:       "4",
			Day:        2,
			Color:      "pink",
			Sex:        "ALL",
			MaxPoints:  1899,
			MaxEntries: 72,
			Price:      10,
		},
		{
			Name:       "5",
			Day:        2,
			Color:      "green",
			Sex:        "ALL",
			MaxPoints:  2199,
			MaxEntries: 72,
			Price:      10,
		},
		{
			Name:       "6",
			Day:        2,
			Color:      "yellow",
			Sex:        "ALL",
			MaxPoints:  1499,
			MaxEntries: 72,
			Price:      9,
		},
		{
			Name:       "7",
			Day:        2,
			Color:      "yellow",
			Sex:        "ALL",
			MaxPoints:  99999, // TS
			MaxEntries: 72,
			Price:      10,
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
