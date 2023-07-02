package public

import (
	"github.com/SuperPingPong/tournoi/internal/auth"
	"github.com/SuperPingPong/tournoi/internal/middlewares"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type API struct {
	db             *gorm.DB
	router         *gin.Engine
	httpClient     HTTPClient
	authMiddleware *jwt.GinJWTMiddleware
}

func NewAPI(db *gorm.DB, r *gin.Engine, client HTTPClient) *API {
	c := &API{
		db:         db,
		router:     r,
		httpClient: client,
	}

	c.setupRouter()
	return c
}

func (api *API) setupRouter() {
	var err error

	authBusiness := auth.NewAuthBusiness(api.db)
	api.authMiddleware, err = authBusiness.AuthMiddleware()
	if err != nil {
		panic(err)
	}

	api.router.Use(middlewares.ErrorHandler())
	api.router.POST("/api/otp", api.SendOTP)
	api.router.POST("/api/login", api.authMiddleware.LoginHandler)
	api.router.GET("/api/logout", api.authMiddleware.LogoutHandler)

	authenticated := api.router.Group("/")
	authenticated.Use(api.authMiddleware.MiddlewareFunc())
	{
		authenticated.GET("/api/members", api.ListMembers)
		authenticated.POST("/api/members", api.CreateMember)
		authenticated.PATCH("/api/members/:id", api.UpdateMember)
	}
}
