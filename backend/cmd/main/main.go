package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"

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

	var mandatoryEnvVars = []string{
		"ADMIN_EMAIL", "EXTERNAL_URL", "JWT_SECRET_KEY", "SENTRY_DSN", "TOKEN_JSON", "CREDENTIALS_JSON",
	}
	for _, envVar := range mandatoryEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s environment variable is empty", envVar)
		}
	}

	sentryDsn := os.Getenv("SENTRY_DSN")
	err = sentry.Init(sentry.ClientOptions{
		Dsn: sentryDsn,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	r := gin.Default()

	public.NewAPI(db, r, &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}, sentryDsn)

	bands := []models.Band{
		{
			Name:           "A",
			Day:            1,
			Color:          "blue",
			SexAllowed:     "ALL",
			MaxPoints:      599,
			MaxEntries:     48,
			Price:          9,
			OnlyCategories: []string{"P", "B1", "B2", "M1", "M2"},
		},
		{
			Name:       "B",
			Day:        1,
			Color:      "blue",
			SexAllowed: "ALL",
			MaxPoints:  1099,
			MaxEntries: 48,
			Price:      9,
		},
		{
			Name:       "C",
			Day:        1,
			Color:      "pink",
			SexAllowed: "ALL",
			MaxPoints:  699,
			MaxEntries: 76,
			Price:      9,
		},
		{
			Name:       "D",
			Day:        1,
			Color:      "pink",
			SexAllowed: "ALL",
			MaxPoints:  1299,
			MaxEntries: 76,
			Price:      10,
		},
		{
			Name:       "E",
			Day:        1,
			Color:      "yellow",
			SexAllowed: "ALL",
			MaxPoints:  899,
			MaxEntries: 76,
			Price:      9,
		},
		{
			Name:       "F",
			Day:        1,
			Color:      "yellow",
			SexAllowed: "ALL",
			MaxPoints:  1499,
			MaxEntries: 76,
			Price:      10,
		},
		{
			Name:       "G",
			Day:        1,
			Color:      "green",
			SexAllowed: "F",
			MaxPoints:  1199,
			MaxEntries: 36,
			Price:      9,
		},
		{
			Name:       "H",
			Day:        1,
			Color:      "green",
			SexAllowed: "ALL",
			MaxPoints:  1699,
			MaxEntries: 76,
			Price:      10,
		},
		{
			Name:       "1",
			Day:        2,
			Color:      "blue",
			SexAllowed: "ALL",
			MaxPoints:  799,
			MaxEntries: 96,
			Price:      9,
		},
		{
			Name:       "2",
			Day:        2,
			Color:      "pink",
			SexAllowed: "ALL",
			MaxPoints:  1399,
			MaxEntries: 82,
			Price:      10,
		},
		{
			Name:       "3",
			Day:        2,
			Color:      "pink",
			SexAllowed: "ALL",
			MaxPoints:  999,
			MaxEntries: 82,
			Price:      9,
		},
		{
			Name:       "4",
			Day:        2,
			Color:      "yellow",
			SexAllowed: "ALL",
			MaxPoints:  599,
			MaxEntries: 82,
			Price:      9,
		},
		{
			Name:       "5",
			Day:        2,
			Color:      "yellow",
			SexAllowed: "ALL",
			MaxPoints:  1599,
			MaxEntries: 82,
			Price:      10,
		},
		{
			Name:       "6",
			Day:        2,
			Color:      "green",
			SexAllowed: "ALL",
			MaxPoints:  1199,
			MaxEntries: 82,
			Price:      9,
		},
		{
			Name:       "7",
			Day:        2,
			Color:      "green",
			SexAllowed: "ALL",
			MaxPoints:  1899,
			MaxEntries: 82,
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

	_, err = public.GetGmailService()
	if err != nil {
		panic(err)
	}

	err = r.Run("0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
}
