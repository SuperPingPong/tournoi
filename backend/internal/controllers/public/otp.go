package public

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/SuperPingPong/tournoi/internal/models"
	"github.com/gin-gonic/gin/binding"

	"github.com/gin-gonic/gin"
	_ "golang.org/x/oauth2/google"
)

const otpExpirationDelay = 10 * time.Minute

type SendOTPInput struct {
	Email string `binding:"required,email"`
}

func (api *API) SendOTP(ctx *gin.Context) {
	var input SendOTPInput

	err := ctx.ShouldBindBodyWith(&input, binding.JSON)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	// Convert the email to lowercase
	input.Email = strings.ToLower(input.Email)

	// Check if there is a valid otp for the input email in database
	var existingOTP models.OTP
	err = api.db.
		Model(&models.OTP{}).
		Where("email = ?", input.Email).
		Where("email = ? AND expires_at > ? AND deleted_at IS NULL", input.Email, time.Now()).
		Order("created_at DESC").
		First(&existingOTP).
		Error

	var password string
	if err == nil { // OTP can be found
		password = existingOTP.Secret
		ctx.Status(http.StatusOK)
		return
	}

	password, err = generatePassword()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to generate OTP: %w", err))
		return
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

	err = sendEmailHTMLOTP(input.Email, password)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to send email: %w", err))
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
