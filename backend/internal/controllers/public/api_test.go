package public

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SuperPingPong/tournoi/internal/auth"
	"github.com/SuperPingPong/tournoi/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type testEnv struct {
	ctx      *gin.Context
	api      *API
	db       *gorm.DB
	jwt      string
	user     *models.User
	teardown func()
}

func getTestEnv(t *testing.T) testEnv {
	// Init DB transaction
	db, err := models.ConnectDatabase()
	if err != nil {
		panic(err)
	}
	tx := db.Begin(&sql.TxOptions{})

	// Init API
	recorder := httptest.NewRecorder()
	ctx, r := gin.CreateTestContext(recorder)

	mockHTTPClient := NewMockHTTPClient(t)
	api := NewAPI(tx, r, mockHTTPClient)

	// Create OTP
	otp := models.OTP{
		Email:     "test@example.com",
		Secret:    "123456",
		ExpiresAt: time.Now().Add(otpExpirationDelay),
	}
	err = tx.Create(&otp).Error
	if err != nil {
		panic(err)
	}

	// Login
	body, err := json.Marshal(auth.LoginRequest{Email: otp.Email, Secret: otp.Secret})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "/login", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	api.router.ServeHTTP(recorder, req)

	// Get JWT
	type loginResponse struct {
		Token string `json:"token"`
	}
	var response loginResponse
	err = json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		panic(err)
	}

	// Get test user
	var user models.User
	err = tx.Where(&models.User{Email: otp.Email}).First(&user).Error
	if err != nil {
		panic(err)
	}

	return testEnv{
		ctx:  ctx,
		api:  api,
		db:   tx,
		jwt:  response.Token,
		user: &user,
		teardown: func() {
			tx.Rollback()
		},
	}
}

func performRequest(method, target string, body io.Reader, headers map[string]string, router *gin.Engine) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, body)
	w := httptest.NewRecorder()

	for k, v := range headers {
		r.Header.Set(k, v)
	}

	router.ServeHTTP(w, r)
	return w
}
