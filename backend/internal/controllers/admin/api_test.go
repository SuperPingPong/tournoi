package admin

import (
	"net/http/httptest"

	"github.com/SuperPingPong/tournoi/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type testEnv struct {
	ctx      *gin.Context
	c        *API
	w        *httptest.ResponseRecorder
	db       *gorm.DB
	teardown func()
}

func getTestEnv() testEnv {
	db, err := models.ConnectDatabase()
	if err != nil {
		panic(err)
	}

	w := httptest.NewRecorder()
	ctx, r := gin.CreateTestContext(w)
	c := NewAPI(db, r)

	return testEnv{
		ctx: ctx,
		c:   c,
		w:   w,
		db:  db,
		teardown: func() {

		},
	}
}
