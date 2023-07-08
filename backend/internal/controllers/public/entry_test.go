package public

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/SuperPingPong/tournoi/internal/models"
	"github.com/stretchr/testify/require"
)

func TestListAvailableBands(t *testing.T) {
	type listBandAvailabilitiesResponse struct {
		Bands []BandAvailability
	}
	t.Run("Success", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		bands := []models.Band{
			{
				Name:       "S",
				Day:        1,
				Sex:        models.BandSex_ALL,
				MaxEntries: 3,
				MaxPoints:  100,
			},
			{
				Name:       "T",
				Day:        1,
				Sex:        models.BandSex_M,
				MaxEntries: 1,
				MaxPoints:  200,
			},
			{
				Name:       "U",
				Day:        2,
				Sex:        models.BandSex_F,
				MaxEntries: 2,
				MaxPoints:  200,
			},
			{
				Name:       "V",
				Day:        2,
				Sex:        models.BandSex_ALL,
				MaxEntries: 1,
				MaxPoints:  300,
			},
		}
		require.NoError(t, env.db.Create(&bands).Error)

		members := []models.Member{
			{
				FirstName:  "John",
				LastName:   "Doe",
				Sex:        "M",
				PermitID:   "000000",
				Points:     99.0,
				Category:   "V2",
				ClubName:   "Jane Club",
				PermitType: "T",
				UserID:     env.user.ID,
			},
			{
				FirstName:  "Jane",
				LastName:   "Doe",
				Sex:        "F",
				PermitID:   "000001",
				Points:     199.0,
				Category:   "B1",
				ClubName:   "Jane Club",
				PermitType: "P",
				UserID:     env.user.ID,
			},
			{
				FirstName:  "Joe",
				LastName:   "Dohn",
				Sex:        "M",
				PermitID:   "000002",
				Points:     299.0,
				Category:   "B1",
				ClubName:   "Jane Club",
				PermitType: "P",
				UserID:     env.user.ID,
			},
		}
		require.NoError(t, env.db.Create(&members).Error)

		// John lists availabilities
		url := "/api/members/%s/band-availabilities"
		res := performRequest("GET", fmt.Sprintf(url, members[0].ID), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var got listBandAvailabilitiesResponse
		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Bands, 3)
		require.Equal(t, BandAvailability{
			Band:      bands[0],
			Available: bands[0].MaxEntries,
			Waiting:   0,
		}, got.Bands[0])
		require.Equal(t, BandAvailability{
			Band:      bands[1],
			Available: bands[1].MaxEntries,
			Waiting:   0,
		}, got.Bands[1])
		require.Equal(t, BandAvailability{
			Band:      bands[3],
			Available: bands[3].MaxEntries,
			Waiting:   0,
		}, got.Bands[2])

		// Jane lists availabilities
		res = performRequest("GET", fmt.Sprintf(url, members[1].ID), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Bands, 2)
		require.Equal(t, BandAvailability{
			Band:      bands[2],
			Available: bands[2].MaxEntries,
			Waiting:   0,
		}, got.Bands[0])
		require.Equal(t, BandAvailability{
			Band:      bands[3],
			Available: bands[3].MaxEntries - 1, // John locked an entry
			Waiting:   0,
		}, got.Bands[1])

		// Joe lists availabilities
		res = performRequest("GET", fmt.Sprintf(url, members[2].ID), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Bands, 1)
		require.Equal(t, BandAvailability{
			Band:      bands[3],
			Available: 0, // John still has the lock
			Waiting:   1, // Jane is waiting
		}, got.Bands[0])

		// John lists his availabilities again
		res = performRequest("GET", fmt.Sprintf(url, members[0].ID), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Bands, 3)
		require.Equal(t, BandAvailability{
			Band:      bands[0],
			Available: bands[0].MaxEntries,
			Waiting:   0,
		}, got.Bands[0])
		require.Equal(t, BandAvailability{
			Band:      bands[1],
			Available: bands[1].MaxEntries,
			Waiting:   0,
		}, got.Bands[1])
		require.Equal(t, BandAvailability{
			Band:      bands[3],
			Available: 0, // Jane has the lock now that John refreshed
			Waiting:   1, // Joe is waiting
		}, got.Bands[2])
	})
	t.Run("WrongUser", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		user := &models.User{
			Email: "hdupont@example.com",
			Members: []models.Member{
				{
					FirstName: "Hervé",
					LastName:  "Dupont",
					Sex:       "M",
					PermitID:  "000003",
				},
			},
		}
		env.db.Create(user)

		url := "/api/members/%s/band-availabilities"
		res := performRequest("GET", fmt.Sprintf(url, user.Members[0].ID), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusNotFound, res.Code)
	})
}
