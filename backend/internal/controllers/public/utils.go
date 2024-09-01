package public

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"

	"github.com/SuperPingPong/tournoi/internal/auth"
	"github.com/SuperPingPong/tournoi/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Paginate(page int, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func ExtractUserFromContext(ctx *gin.Context) (*models.User, error) {
	userValue, ok := ctx.Get(auth.IdentityKey)
	if !ok {
		return nil, fmt.Errorf("failed to get current user")
	}

	user, ok := userValue.(*models.User)
	if !ok {
		return nil, fmt.Errorf("failed to extract current user from context")
	}

	return user, nil
}

func FilterByUserID(user *models.User) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if user.IsAdmin {
			return db
		}
		return db.Where("user_id = ?", user.ID)
	}
}

func loadTokenFromEnv() (*oauth2.Token, error) {
	file := os.Getenv("TOKEN_JSON")
	token := &oauth2.Token{}
	err := json.Unmarshal([]byte(file), token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func saveTokenToEnv(token *oauth2.Token) error {
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return err
	}

	err = os.Setenv("TOKEN_JSON", string(tokenJson))
	if err != nil {
		return err
	}

	return nil
}

func getClient(config *oauth2.Config) (*http.Client, error) {
	token, err := loadTokenFromEnv()
	if err != nil {
		return nil, err
	}

	if !token.Valid() {
		tokenSource := config.TokenSource(oauth2.NoContext, token)
		newToken, err := tokenSource.Token()
		if err != nil {
			return nil, err
		}

		// Save the updated token
		err = saveTokenToEnv(newToken)
		if err != nil {
			return nil, err
		}

		token = newToken
	}

	return config.Client(oauth2.NoContext, token), nil
}

func GetGmailService() (*gmail.Service, error) {

	// Load client credentials
	clientSecretData := os.Getenv("CREDENTIALS_JSON")

	config, err := google.ConfigFromJSON([]byte(clientSecretData), gmail.MailGoogleComScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// Obtain an OAuth client
	client, err := getClient(config)
	if err != nil {
		log.Fatalf("Unable to get OAuth client: %v", err)
	}

	// Create a Gmail service
	var service *gmail.Service
	service, err = gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to create Gmail service: %v", err)
	}

	return service, err
}

/*
func sendEmailOTP(to string, code string) error {
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
*/

func sendEmailHTMLOTP(to string, code string) error {
	service, err := GetGmailService()
	if err != nil {
		return fmt.Errorf("failed to get Gmail service: %v", err)
	}

	// Read the HTML content from the file
	htmlContent, err := ioutil.ReadFile("email_templates/otp_code.html")
	if err != nil {
		return fmt.Errorf("failed to read email HTML file: %v", err)
	}

	// Replace the placeholder with the otp code
	replacedContent := strings.Replace(string(htmlContent), "OTP_CODE", code, -1)

	// Set up the email message
	message := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(
			fmt.Sprintf("To: %s\r\nSubject: OTP %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n\r\n%s", to, code, replacedContent)),
		),
	}

	_, err = service.Users.Messages.Send("me", message).Do()
	if err != nil {
		return err
	}

	return nil
}

func sendEmailHTML(to string, lastName string, firstName string) error {
	service, err := GetGmailService()
	if err != nil {
		return fmt.Errorf("failed to get Gmail service: %v", err)
	}

	// Read the HTML content from the file
	htmlContent, err := ioutil.ReadFile("email_templates/register_confirm.html")
	if err != nil {
		return fmt.Errorf("failed to read email HTML file: %v", err)
	}

	// Replace the placeholder with the environment variable
	externalURL := os.Getenv("EXTERNAL_URL")
	if externalURL == "" {
		return fmt.Errorf("EXTERNAL_URL environment variable not set")
	}
	replacedContent := strings.Replace(string(htmlContent), "EXTERNAL_URL", externalURL, -1)

	// Escape the first name and last name to ensure proper encoding
	escapedFirstName := html.EscapeString(firstName)
	escapedLastName := html.EscapeString(lastName)

	// Set up the email message
	subject := fmt.Sprintf("Confirmation inscription %s %s Tournoi de Lognes", escapedLastName, escapedFirstName)
	encodedSubject := encodeHeader(subject)
	message := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(
			fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n\r\n%s", to, encodedSubject, replacedContent)),
		),
	}

	_, err = service.Users.Messages.Send("me", message).Do()
	if err != nil {
		return err
	}

	return nil
}

// encodeHeader encodes special characters in the given header string using MIME encoding
func encodeHeader(header string) string {
	encoded := mime.QEncoding.Encode("utf-8", header)
	return encoded
}
