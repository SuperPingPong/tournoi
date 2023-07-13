package public

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/getsentry/sentry-go"
)

//go:generate mockery --name HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type FFTTPlayer struct {
	LastName   string  `json:"nom"`
	FirstName  string  `json:"prenom"`
	PermitID   string  `json:"licence"`
	Sex        string  `json:"sexe"`
	Points     float64 `json:"point"`
	Category   string  `json:"cat,omitempty"`
	ClubName   string  `json:"nomclub"`
	PermitType string  `json:"type,omitempty"`
}

func (api *API) GetFFTTPlayerData(permitID string) (*FFTTPlayer, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://fftt.dafunker.com/v1/joueur/%s", permitID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := api.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch player %s from FFTT: %w", permitID, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response for player %s from FFTT: %w", permitID, err)
	}

	var data FFTTPlayer
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response for player %s from FFTT: %w", permitID, err)
	}

	return &data, nil
}

func (api *API) GetFFTTPlayer(ctx *gin.Context) {
	permitID := ctx.Param("id")

	player, err := api.GetFFTTPlayerData(permitID)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("failed to get FFTT player: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get FFTT player: %w", err))
	}
	if len(player.LastName) == 0 && len(player.FirstName) == 0 {
		sentry.CaptureException(fmt.Errorf("FFTT player %s not found", permitID))
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("FFTT player %s not found", permitID))
		return
	}

	ctx.JSON(http.StatusOK, &player)
}

type PlayerXML struct {
	LastName  string  `xml:"nom"`
	FirstName string  `xml:"prenom"`
	ClubName  string  `xml:"nclub"`
	Points    float64 `xml:"points"`
	PermitID  string  `xml:"licence"`
	Sex       string  `xml:"sexe"`
}

type PlayersXML struct {
	Players []PlayerXML `xml:"joueur"`
}

type SearchFFTTPlayersInput struct {
	Surname string
	Name    string
}

func (api *API) SearchFFTTPlayers(ctx *gin.Context) {
	var input SearchFFTTPlayersInput
	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("invalid input: %w", err))
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://fftt.dafunker.com/v1//proxy/xml_liste_joueur_o.php", nil)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("failed to build request: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to build request: %w", err))
		return
	}

	queryParams := map[string]string{
		"nom":    input.Surname,
		"prenom": input.Name,
	}
	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	if err != nil {
		sentry.CaptureException(fmt.Errorf("failed to build request query parameters: %w", err))
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to build request query parameters: %w", err))
		return
	}

	resp, err := api.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		sentry.CaptureException(fmt.Errorf("failed to search FFTT player: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to search FFTT player: %w", err))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("failed to parse search response from FFTT: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to parse search response from FFTT: %w", err))
		return
	}

	var data PlayersXML
	xmlString := strings.Replace(string(body), "encoding=\"ISO-8859-1\"", "encoding=\"UTF-8\"", 1)
	xmlBytes := []byte(xmlString)
	err = xml.Unmarshal(xmlBytes, &data)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("failed to unmarshal search response from FFTT: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to unmarshal search response from FFTT: %w", err))
		return
	}

	var ffttPlayers []FFTTPlayer
	for _, player := range data.Players {
		if len(ffttPlayers) == 10 {
			break
		}
		if player.ClubName == "" {
			continue
		}

		ffttPlayers = append(ffttPlayers, FFTTPlayer{
			LastName:  strings.TrimSpace(player.LastName),
			FirstName: strings.TrimSpace(player.FirstName),
			PermitID:  player.PermitID,
			Points:    player.Points,
			ClubName:  player.ClubName,
			Sex:       player.Sex,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"players": ffttPlayers})
}
