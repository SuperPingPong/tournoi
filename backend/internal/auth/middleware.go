package auth

import (
	"errors"
	"time"

	"github.com/SuperPingPong/tournoi/internal/models"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var IdentityKey = "uid"

func (a *AuthBusiness) AuthMiddleware() (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour * 24 * 365,
		MaxRefresh:  time.Hour * 24 * 365,
		IdentityKey: IdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if user, ok := data.(*models.User); ok {
				return jwt.MapClaims{
					IdentityKey: user.ID,
				}
			}
			return jwt.MapClaims{}
		},
		Authenticator: func(ctx *gin.Context) (interface{}, error) {
			var loginRequest LoginRequest
			if err := ctx.ShouldBindJSON(&loginRequest); err != nil {
				return "", errors.New("authentication required")
			}

			user, err := a.Login(loginRequest.Email, loginRequest.Secret)
			if err != nil {
				return "", errors.New("authentication failed")
			}

			return user, nil
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)

			var user models.User
			a.db.First(&user, uuid.MustParse(claims[IdentityKey].(string)))
			return &user
		},
		Authorizator: func(data interface{}, ctx *gin.Context) bool {
			if _, ok := data.(*models.User); ok {
				return true
			}
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:    "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:  "Bearer",
		TimeFunc:       time.Now,
		SendCookie:     true,
		CookieHTTPOnly: true,
	})

	if err != nil {
		return nil, err
	}

	return authMiddleware, nil
}
