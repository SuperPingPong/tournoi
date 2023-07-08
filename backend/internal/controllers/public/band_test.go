package public

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/SuperPingPong/tournoi/internal/models"
	"github.com/stretchr/testify/require"
)

func TestSetMemberEntries(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		bands := []models.Band{
			{
				Name: "S",
			},
			{
				Name: "T",
			},
			{
				Name: "U",
			},
		}
		require.NoError(t, env.db.Create(&bands).Error)

		member := models.Member{
			FirstName: "John",
			LastName:  "Doe",
			Sex:       "M",
			PermitID:  "000001",
			UserID:    env.user.ID,
		}
		require.NoError(t, env.db.Create(&member).Error)

		memberBands := []models.Entry{
			{
				MemberID: member.ID,
				BandID:   bands[0].ID,
			},
			{
				MemberID: member.ID,
				BandID:   bands[1].ID,
			},
		}
		require.NoError(t, env.db.Create(&memberBands).Error)

		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		data := map[string][]string{
			"BandIDs": {
				bands[0].ID.String(),
				bands[2].ID.String(),
			},
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.Code)

		var updatedEntries []models.Entry
		require.NoError(t, env.db.Where(&models.Entry{MemberID: member.ID}).Find(&updatedEntries).Error)
		require.Len(t, updatedEntries, 2)
		require.Equal(t, member.ID, updatedEntries[0].MemberID)
		require.Equal(t, bands[0].ID, updatedEntries[0].BandID)
		require.Equal(t, member.ID, updatedEntries[1].MemberID)
		require.Equal(t, bands[2].ID, updatedEntries[1].BandID)

		var deletedEntries []models.Entry
		require.NoError(t, env.db.Unscoped().Where(&models.Entry{BandID: bands[1].ID, MemberID: member.ID}).Find(&deletedEntries).Error)
		require.Len(t, deletedEntries, 1)
		require.Equal(t, member.ID, deletedEntries[0].MemberID)
		require.Equal(t, bands[1].ID, deletedEntries[0].BandID)
		require.True(t, deletedEntries[0].DeletedAt.Valid)
	})
}

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

func TestListAvailableBands(t *testing.T) {
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
