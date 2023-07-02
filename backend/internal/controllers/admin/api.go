package admin

import (
	"github.com/SuperPingPong/tournoi/internal/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type API struct {
	db     *gorm.DB
	router *gin.Engine
}

func NewAPI(db *gorm.DB, r *gin.Engine) *API {
	c := &API{
		db:     db,
		router: r,
	}

	c.setupRouter()
	return c
}

func (api *API) setupRouter() {
	api.router.Use(middlewares.ErrorHandler())

	api.router.GET("/bands/:id", api.GetBand)
	api.router.POST("/bands", api.CreateBand)
	api.router.PATCH("/bands/:id", api.UpdateBand)
}
