package public

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"google.golang.org/api/gmail/v1"
	"math/big"
	"net/http"
	"time"

	"github.com/SuperPingPong/tournoi/internal/models"

	"github.com/gin-gonic/gin"
	_ "golang.org/x/oauth2/google"
)

const otpExpirationDelay = 10 * time.Minute

func sendEmail(to string, code string) error {
	service, err := GetGmailService()

	// Set up the email message
	message := &gmail.Message{
		Raw: base64.RawURLEncoding.EncodeToString([]byte(
			fmt.Sprintf("To: %s\r\nSubject: OTP %s Tournoi de Lognes\r\n\r\nVoici votre code de vérification OTP: %s", to, code, code)),
		),
	}

	_, err = service.Users.Messages.Send("me", message).Do()
	if err != nil {
		return err
	}

	return nil
}

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

	err = sendEmail(input.Email, password)
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
