package public

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetFFTTPlayer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		firstName := "Jean"
		lastName := "Pierre"
		permitID := "123456"
		sex := "M"
		point := 801.0
		category := "S"
		clubName := "Caillouville"
		permitType := "T"

		expectedFFTTReq, err := http.NewRequest(http.MethodGet, "https://fftt.dafunker.com/v1/joueur/"+permitID, nil)
		mockFFTTRes := fmt.Sprintf(`{"nom":"%s","prenom":"%s","licence":"%s","sexe":"%s","point":%f,"cat":"%s","nomclub":"%s","type":"%s"}`, lastName, firstName, permitID, sex, point, category, clubName, permitType)
		r := io.NopCloser(bytes.NewReader([]byte(mockFFTTRes)))
		env.api.httpClient.(*MockHTTPClient).EXPECT().Do(expectedFFTTReq).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       r,
		}, nil)

		res := performRequest("GET", "/api/players/"+permitID, nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var player FFTTPlayer
		require.Equal(t, http.StatusOK, res.Code)
		err = json.NewDecoder(res.Body).Decode(&player)
		require.NoError(t, err)
		require.Equal(t, firstName, player.FirstName)
		require.Equal(t, lastName, player.LastName)
		require.Equal(t, permitID, player.PermitID)
		require.Equal(t, sex, player.Sex)
		require.Equal(t, point, player.Points)
		require.Equal(t, category, player.Category)
		require.Equal(t, clubName, player.ClubName)
		require.Equal(t, permitType, player.PermitType)
	})
	t.Run("NotFound", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		permitID := "invalid"

		expectedFFTTReq, err := http.NewRequest(http.MethodGet, "https://fftt.dafunker.com/v1/joueur/"+permitID, nil)
		mockFFTTRes := fmt.Sprintf(`{"nom":"","prenom":"","licence":"%s","sexe":"","point":%d,"cat":"","nomclub":"","type":""}`, permitID, 0)
		r := io.NopCloser(bytes.NewReader([]byte(mockFFTTRes)))
		env.api.httpClient.(*MockHTTPClient).EXPECT().Do(expectedFFTTReq).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       r,
		}, nil)

		res := performRequest("GET", "/api/players/"+permitID, nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var actual map[string]string
		err = json.NewDecoder(res.Body).Decode(&actual)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, res.Code)
		require.Equal(t, fmt.Sprintf("FFTT player %s not found", permitID), actual["error"])
	})
}
