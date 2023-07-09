package public

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/SuperPingPong/tournoi/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSetMemberEntries(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		bands := []models.Band{
			{
				Name:      "S",
				Sex:       models.BandSex_M,
				MaxPoints: 799,
			},
			{
				Name:      "T",
				Sex:       models.BandSex_M,
				MaxPoints: 999,
			},
			{
				Name:      "U",
				Sex:       models.BandSex_ALL,
				MaxPoints: 999,
			},
			{
				Name:      "V",
				Sex:       models.BandSex_F,
				MaxPoints: 1199,
			},
		}
		require.NoError(t, env.db.Create(&bands).Error)

		member := models.Member{
			FirstName: "John",
			LastName:  "Doe",
			Sex:       "M",
			PermitID:  "000000",
			Points:    700,
			UserID:    env.user.ID,
		}
		require.NoError(t, env.db.Create(&member).Error)

		sessionID := uuid.New()
		entries := []models.Entry{
			{
				MemberID:  member.ID,
				BandID:    bands[0].ID,
				Confirmed: true,
				CreatedAt: time.Now().Add(1 * time.Second),
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
			{
				MemberID:  member.ID,
				BandID:    bands[1].ID,
				Confirmed: true,
				CreatedAt: time.Now().Add(2 * time.Second),
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
			{
				MemberID:  member.ID,
				BandID:    bands[2].ID,
				Confirmed: false,
				CreatedAt: time.Now().Add(3 * time.Second),
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
			{
				MemberID:  member.ID,
				BandID:    bands[3].ID,
				Confirmed: false,
				CreatedAt: time.Now().Add(4 * time.Second),
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
		}
		require.NoError(t, env.db.Create(&entries).Error)

		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		data := map[string]interface{}{
			"BandIDs": []string{
				bands[1].ID.String(),
				bands[2].ID.String(),
			},
			"SessionID": sessionID,
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)

		var updatedEntries []models.Entry
		require.NoError(t, env.db.Where(&models.Entry{MemberID: member.ID, Confirmed: true}).Order("created_at ASC").Find(&updatedEntries).Error)
		require.Len(t, updatedEntries, 2)
		require.Equal(t, entries[1].ID, updatedEntries[0].ID)
		require.Equal(t, entries[1].BandID, updatedEntries[0].BandID)
		require.Equal(t, entries[1].MemberID, updatedEntries[0].MemberID)
		require.Equal(t, entries[1].SessionID, updatedEntries[0].SessionID)
		require.True(t, updatedEntries[0].Confirmed)
		require.Equal(t, entries[2].ID, updatedEntries[1].ID)
		require.Equal(t, entries[2].BandID, updatedEntries[1].BandID)
		require.Equal(t, entries[2].MemberID, updatedEntries[1].MemberID)
		require.Equal(t, entries[2].SessionID, updatedEntries[1].SessionID)
		require.True(t, updatedEntries[1].Confirmed)

		var deletedEntries []models.Entry
		require.NoError(t, env.db.Unscoped().Where(&models.Entry{BandID: bands[0].ID, MemberID: member.ID}).Find(&deletedEntries).Error)
		require.Len(t, deletedEntries, 1)
		require.Equal(t, member.ID, deletedEntries[0].MemberID)
		require.Equal(t, bands[0].ID, deletedEntries[0].BandID)
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
