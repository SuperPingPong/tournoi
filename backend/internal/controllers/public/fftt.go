package public

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

//go:generate mockery --name HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type FFTTMemberData struct {
	Nom     string  `json:"nom"`
	Prenom  string  `json:"prenom"`
	Licence string  `json:"licence"`
	Sexe    string  `json:"sexe"`
	Point   float64 `json:"point"`
}

func (api *API) GetFFTTMemberData(permitID string) (*FFTTMemberData, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://fftt.dafunker.com/v1/joueur/%s", permitID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make http request to FFTT: %w", err)
	}

	resp, err := api.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch member %s from FFTT: %w", permitID, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response for member %s from FFTT: %w", permitID, err)
	}

	var data FFTTMemberData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response for member %s from FFTT: %w", permitID, err)
	}

	if len(data.Nom) == 0 && len(data.Prenom) == 0 {
		return nil, fmt.Errorf("failed to find valid member %s from FFTT", permitID)
	}

	return &data, nil
}
