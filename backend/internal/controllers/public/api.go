package public

import (
	"github.com/SuperPingPong/tournoi/internal/auth"
	"github.com/SuperPingPong/tournoi/internal/middlewares"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/getsentry/sentry-go"
)

type API struct {
	db             *gorm.DB
	router         *gin.Engine
	httpClient     HTTPClient
	authMiddleware *jwt.GinJWTMiddleware
}

func NewAPI(db *gorm.DB, r *gin.Engine, client HTTPClient, sentryDSN string) *API {
	// Initialize Sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn: sentryDSN,
	})
	if err != nil {
		panic(err)
	}

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
	api.router.GET("/api/players/:id", api.GetFFTTPlayer)
	api.router.POST("/api/players", api.SearchFFTTPlayers)

	authenticated := api.router.Group("/api")
	authenticated.Use(api.authMiddleware.MiddlewareFunc())
	{
		authenticated.GET("/members", api.ListMembers)
		authenticated.GET("/members/:id", api.GetMember)
		authenticated.POST("/members", api.CreateMember)
		authenticated.DELETE("/members/:id", api.DeleteMember)
		authenticated.GET("/members/:id/get-entries-history", api.GetMemberEntriesHistory)
		authenticated.POST("/members/:id/set-entries", api.SetMemberEntries)
		authenticated.GET("/members/:id/band-availabilities", api.ListBandAvailabilities)
		authenticated.GET("/bands", api.ListBands)
		authenticated.POST("/check-auth", api.CheckAuth)
	}
}
