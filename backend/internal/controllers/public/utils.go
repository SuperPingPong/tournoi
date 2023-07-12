package public

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

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

const (
	TokenFile = "token.json"
)

func loadTokenFromFile() (*oauth2.Token, error) {
	file, err := ioutil.ReadFile(TokenFile)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{}
	err = json.Unmarshal(file, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func saveTokenToFile(token *oauth2.Token) error {
	file, err := os.OpenFile(TokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	fileBytes, err := json.Marshal(token)
	if err != nil {
		return err
	}

	_, err = file.Write(fileBytes)
	if err != nil {
		return err
	}

	return nil
}

func getClient(config *oauth2.Config) (*http.Client, error) {
	token, err := loadTokenFromFile()
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
		err = saveTokenToFile(newToken)
		if err != nil {
			return nil, err
		}

		token = newToken
	}

	return config.Client(oauth2.NoContext, token), nil
}

func GetGmailService() (*gmail.Service, error) {
	const (
		ClientSecretFile = "credentials.json"
		TokenFile        = "token.json"
	)

	// Load client credentials
	var clientSecretData []byte
	clientSecretData, err := ioutil.ReadFile(ClientSecretFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(clientSecretData, gmail.MailGoogleComScope)
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

func sendEmailHTML(to string, lastName string, firstName string) error {
	service, err := GetGmailService()
	if err != nil {
		return fmt.Errorf("failed to get Gmail service: %v", err)
	}

	// Read the HTML content from the file
	htmlContent, err := ioutil.ReadFile("email_template.html")
	if err != nil {
		return fmt.Errorf("failed to read email HTML file: %v", err)
	}

	// Replace the placeholder with the environment variable
	externalURL := os.Getenv("EXTERNAL_URL")
	if externalURL == "" {
		return fmt.Errorf("EXTERNAL_URL environment variable not set")
	}
	replacedContent := strings.Replace(string(htmlContent), "EXTERNAL_URL", externalURL, -1)

	// Set up the email message
	message := &gmail.Message{
		Raw: base64.RawURLEncoding.EncodeToString([]byte(
			fmt.Sprintf("To: %s\r\nSubject: Confirmation inscription %s %s Tournoi de Lognes\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n\r\n%s", to, lastName, firstName, replacedContent)),
		),
	}

	_, err = service.Users.Messages.Send("me", message).Do()
	if err != nil {
		return err
	}

	return nil
}
