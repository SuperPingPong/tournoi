package public

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/SuperPingPong/tournoi/internal/models"
	"github.com/stretchr/testify/require"
)

func TestListBands(t *testing.T) {
	type listBandsResponse struct {
		Bands []models.Band
	}
	t.Run("Success", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		bands := []models.Band{
			{
				Name: "S",
				Day:  1,
			},
			{
				Name: "T",
				Day:  1,
			},
			{
				Name: "U",
				Day:  2,
			},
		}
		require.NoError(t, env.db.Create(&bands).Error)

		var got listBandsResponse

		// No filter
		res := performRequest("GET", "/api/bands", nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		err := json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.Len(t, got.Bands, 3)
		require.Equal(t, bands[0], got.Bands[0])
		require.Equal(t, bands[1], got.Bands[1])
		require.Equal(t, bands[2], got.Bands[2])

		// Filtered by day=1
		res = performRequest("GET", "/api/bands?day=1", nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		err = json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.Len(t, got.Bands, 2)
		require.Equal(t, bands[0], got.Bands[0])
		require.Equal(t, bands[1], got.Bands[1])
	})
}
