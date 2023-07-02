package public

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/SuperPingPong/tournoi/internal/models"

	"github.com/gin-gonic/gin"
)

const otpExpirationDelay = 10 * time.Minute

type SendOTPInput struct {
	Email string `binding:"required,email"`
}

func (api *API) SendOTP(ctx *gin.Context) {
	var input SendOTPInput

	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	password, err := generatePassword()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to generate OTP: %w", err))
	}
	otp := models.OTP{
		Email:     input.Email,
		Secret:    password,
		ExpiresAt: time.Now().Add(otpExpirationDelay),
	}
	err = api.db.Create(&otp).Error
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create OTP: %w", err))
		return
	}

	ctx.Status(http.StatusOK)
}

func generatePassword() (string, error) {
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
