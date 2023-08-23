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
				MaxPoints:  99,
			},
			{
				Name:       "T",
				Day:        1,
				Sex:        models.BandSex_M,
				MaxEntries: 1,
				MaxPoints:  199,
			},
			{
				Name:       "U",
				Day:        2,
				Sex:        models.BandSex_F,
				MaxEntries: 2,
				MaxPoints:  199,
			},
			{
				Name:       "V",
				Day:        2,
				Sex:        models.BandSex_ALL,
				MaxEntries: 1,
				MaxPoints:  299,
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
		var createdEntries []models.Entry
		require.NoError(t, env.db.Where(&models.Entry{MemberID: members[0].ID}).Order("created_at ASC").Find(&createdEntries).Error)
		require.Len(t, createdEntries, 3)
		sessionID := createdEntries[0].SessionID
		for _, entry := range createdEntries {
			require.Equal(t, sessionID, entry.SessionID)
			require.True(t, entry.CreatedBy.Valid)
			require.Equal(t, env.user.ID, entry.CreatedBy.UUID)
			require.False(t, entry.Confirmed)
			require.False(t, entry.ConfirmedBy.Valid)
			require.False(t, entry.DeletedAt.Valid)
			require.False(t, entry.DeletedBy.Valid)
		}

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
	t.Run("SuccessWithExistingConfirmedEntries", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		bands := []models.Band{
			{
				Name:       "S",
				Day:        1,
				Sex:        models.BandSex_ALL,
				MaxEntries: 3,
				MaxPoints:  99,
			},
			{
				Name:       "T",
				Day:        1,
				Sex:        models.BandSex_M,
				MaxEntries: 1,
				MaxPoints:  199,
			},
			{
				Name:       "V",
				Day:        2,
				Sex:        models.BandSex_ALL,
				MaxEntries: 1,
				MaxPoints:  299,
			},
		}
		require.NoError(t, env.db.Create(&bands).Error)

		member := models.Member{
			FirstName:  "John",
			LastName:   "Doe",
			Sex:        "M",
			PermitID:   "000000",
			Points:     99.0,
			Category:   "V2",
			ClubName:   "Jane Club",
			PermitType: "T",
			UserID:     env.user.ID,
		}
		require.NoError(t, env.db.Create(&member).Error)

		sessionID := uuid.New()
		entries := []models.Entry{
			{
				MemberID:    member.ID,
				BandID:      bands[0].ID,
				CreatedAt:   time.Now().Add(-2 * time.Second),
				ExpiresAt:   time.Now().Add(time.Hour),
				Confirmed:   true,
				ConfirmedBy: uuid.NullUUID{UUID: env.user.ID, Valid: true},
				SessionID:   sessionID,
			},
			{
				MemberID:    member.ID,
				BandID:      bands[1].ID,
				CreatedAt:   time.Now().Add(-1 * time.Second),
				ExpiresAt:   time.Now().Add(time.Hour),
				Confirmed:   true,
				ConfirmedBy: uuid.NullUUID{UUID: env.user.ID, Valid: true},
				SessionID:   sessionID,
			},
		}
		require.NoError(t, env.db.Create(&entries).Error)

		// John lists availabilities
		url := "/api/members/%s/band-availabilities"
		res := performRequest("GET", fmt.Sprintf(url, member.ID), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var got listBandAvailabilitiesResponse
		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Bands, 3)
		require.Equal(t, BandAvailability{
			Band:      bands[0],
			Available: bands[0].MaxEntries - 1, // John has a confirmed entry
			Waiting:   0,
		}, got.Bands[0])
		require.Equal(t, BandAvailability{
			Band:      bands[1],
			Available: bands[1].MaxEntries - 1, // John has a confirmed entry
			Waiting:   0,
		}, got.Bands[1])
		require.Equal(t, BandAvailability{
			Band:      bands[2],
			Available: bands[2].MaxEntries, // No existing entry
			Waiting:   0,
		}, got.Bands[2])

		var currentEntries []models.Entry
		require.NoError(t, env.db.Where(&models.Entry{MemberID: member.ID}).Order("created_at ASC").Find(&currentEntries).Error)
		require.Len(t, currentEntries, 3)

		// First two entries have not been updated
		require.Equal(t, entries[0].ID, currentEntries[0].ID)
		require.Equal(t, entries[0].BandID, currentEntries[0].BandID)
		require.True(t, entries[0].CreatedAt.Equal(currentEntries[0].CreatedAt))
		require.Equal(t, entries[0].SessionID, currentEntries[0].SessionID)
		require.True(t, currentEntries[0].Confirmed)
		require.Equal(t, entries[0].ConfirmedBy, currentEntries[0].ConfirmedBy)

		require.Equal(t, entries[1].ID, currentEntries[1].ID)
		require.Equal(t, entries[1].BandID, currentEntries[1].BandID)
		require.True(t, entries[1].CreatedAt.Equal(currentEntries[1].CreatedAt))
		require.Equal(t, entries[1].SessionID, currentEntries[1].SessionID)
		require.True(t, currentEntries[1].Confirmed)
		require.Equal(t, entries[1].ConfirmedBy, currentEntries[1].ConfirmedBy)

		// A lock has been created for the third band
		require.Equal(t, bands[2].ID, currentEntries[2].BandID)
		require.NotEqual(t, sessionID, currentEntries[2].SessionID)
		require.False(t, currentEntries[2].Confirmed)
		require.False(t, currentEntries[2].ConfirmedBy.Valid)
	})
	t.Run("WrongUser", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		user := models.User{
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
		env.db.Create(&user)

		url := "/api/members/%s/band-availabilities"
		res := performRequest("GET", fmt.Sprintf(url, user.Members[0].ID), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusNotFound, res.Code)
	})
}

func TestSetMemberEntries(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		bands := []models.Band{
			{
				Name:      "S",
				Sex:       models.BandSex_M,
				MaxPoints: 799,
				Color:     models.BandColor_GREEN,
				Day:       1,
			},
			{
				Name:      "T",
				Sex:       models.BandSex_M,
				MaxPoints: 999,
				Color:     models.BandColor_BLUE,
				Day:       2,
			},
			{
				Name:      "U",
				Sex:       models.BandSex_ALL,
				MaxPoints: 999,
				Color:     models.BandColor_GREEN,
				Day:       2,
			},
			{
				Name:      "V",
				Sex:       models.BandSex_F,
				MaxPoints: 1199,
				Color:     models.BandColor_PINK,
				Day:       2,
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
				MemberID:    member.ID,
				BandID:      bands[1].ID,
				Confirmed:   true,
				CreatedAt:   time.Now().Add(2 * time.Second),
				ExpiresAt:   time.Now().Add(time.Hour),
				SessionID:   sessionID,
				ConfirmedBy: uuid.NullUUID{UUID: uuid.New(), Valid: true},
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
		// The first entry was already confirmed
		require.Equal(t, entries[1].ID, updatedEntries[0].ID)
		require.Equal(t, entries[1].BandID, updatedEntries[0].BandID)
		require.True(t, entries[1].CreatedAt.Equal(updatedEntries[0].CreatedAt))
		require.Equal(t, entries[1].MemberID, updatedEntries[0].MemberID)
		require.Equal(t, entries[1].SessionID, updatedEntries[0].SessionID)
		require.True(t, updatedEntries[0].Confirmed)
		require.Equal(t, entries[1].ConfirmedBy, updatedEntries[0].ConfirmedBy)
		// The second one has just been confirmed
		require.Equal(t, entries[2].ID, updatedEntries[1].ID)
		require.Equal(t, entries[2].BandID, updatedEntries[1].BandID)
		require.Equal(t, entries[2].MemberID, updatedEntries[1].MemberID)
		require.Equal(t, entries[2].SessionID, updatedEntries[1].SessionID)
		require.True(t, updatedEntries[1].Confirmed)
		require.True(t, updatedEntries[1].ConfirmedBy.Valid)
		require.Equal(t, env.user.ID, updatedEntries[1].ConfirmedBy.UUID)

		var deletedEntries []models.Entry
		require.NoError(t, env.db.Unscoped().Where("deleted_at IS NOT NULL AND member_id = ?", member.ID).Order("created_at ASC").Find(&deletedEntries).Error)
		require.Len(t, deletedEntries, 2)
		require.Equal(t, entries[0].ID, deletedEntries[0].ID)
		require.Equal(t, entries[0].MemberID, deletedEntries[0].MemberID)
		require.Equal(t, entries[0].BandID, deletedEntries[0].BandID)
		require.True(t, deletedEntries[0].DeletedAt.Valid)
		require.True(t, deletedEntries[0].DeletedBy.Valid)
		require.Equal(t, env.user.ID, deletedEntries[0].DeletedBy.UUID)

		require.Equal(t, entries[3].ID, deletedEntries[1].ID)
		require.Equal(t, entries[3].MemberID, deletedEntries[1].MemberID)
		require.Equal(t, entries[3].BandID, deletedEntries[1].BandID)
		require.True(t, deletedEntries[1].DeletedAt.Valid)
		require.True(t, deletedEntries[1].DeletedBy.Valid)
		require.Equal(t, env.user.ID, deletedEntries[1].DeletedBy.UUID)
	})
	t.Run("RemoveEntry", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		bands := []models.Band{
			{
				Name:      "S",
				Sex:       models.BandSex_M,
				MaxPoints: 799,
				Color:     models.BandColor_GREEN,
				Day:       1,
			},
			{
				Name:      "T",
				Sex:       models.BandSex_M,
				MaxPoints: 999,
				Color:     models.BandColor_BLUE,
				Day:       2,
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
				MemberID:    member.ID,
				BandID:      bands[1].ID,
				Confirmed:   true,
				CreatedAt:   time.Now().Add(2 * time.Second),
				ExpiresAt:   time.Now().Add(time.Hour),
				SessionID:   sessionID,
				ConfirmedBy: uuid.NullUUID{UUID: uuid.New(), Valid: true},
			},
		}
		require.NoError(t, env.db.Create(&entries).Error)

		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		data := map[string]interface{}{
			"BandIDs": []string{
				bands[1].ID.String(),
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
		require.Len(t, updatedEntries, 1)
		// Only the second entry is still confirmed
		require.Equal(t, entries[1].ID, updatedEntries[0].ID)
		require.Equal(t, entries[1].BandID, updatedEntries[0].BandID)
		require.Equal(t, entries[1].MemberID, updatedEntries[0].MemberID)
		require.Equal(t, entries[1].SessionID, updatedEntries[0].SessionID)
		require.True(t, entries[1].CreatedAt.Equal(updatedEntries[0].CreatedAt))
		require.Equal(t, entries[1].CreatedBy, updatedEntries[0].CreatedBy)
		require.True(t, updatedEntries[0].Confirmed)
		require.Equal(t, entries[1].ConfirmedBy, updatedEntries[0].ConfirmedBy)

		// The first entry has been deleted
		var deletedEntries []models.Entry
		require.NoError(t, env.db.Unscoped().Where("deleted_at IS NOT NULL AND member_id = ?", member.ID).Order("created_at ASC").Find(&deletedEntries).Error)
		require.Len(t, deletedEntries, 1)
		require.Equal(t, entries[0].ID, deletedEntries[0].ID)
		require.Equal(t, entries[0].MemberID, deletedEntries[0].MemberID)
		require.Equal(t, entries[0].BandID, deletedEntries[0].BandID)
		require.True(t, deletedEntries[0].DeletedAt.Valid)
		require.True(t, deletedEntries[0].DeletedBy.Valid)
		require.Equal(t, env.user.ID, deletedEntries[0].DeletedBy.UUID)
		
		// Delete all entries
		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		data := map[string]interface{}{
			"BandIDs": []string{},
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
		// The second entry has been deleted as well
		require.Len(t, updatedEntries, 0)

		var deletedEntries []models.Entry
		require.NoError(t, env.db.Unscoped().Where("deleted_at IS NOT NULL AND member_id = ?", member.ID).Order("created_at ASC").Find(&deletedEntries).Error)
		require.Len(t, deletedEntries, 2)
	})
	t.Run("LimitPerDayReached", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		bands := []models.Band{
			{
				Name:      "S",
				Sex:       models.BandSex_M,
				MaxPoints: 799,
				Color:     models.BandColor_PINK,
				Day:       1,
			},
			{
				Name:      "T",
				Sex:       models.BandSex_ALL,
				MaxPoints: 999,
				Color:     models.BandColor_PINK,
				Day:       2,
			},
			{
				Name:      "U",
				Sex:       models.BandSex_M,
				MaxPoints: 999,
				Color:     models.BandColor_GREEN,
				Day:       2,
			},
			{
				Name:      "V",
				Sex:       models.BandSex_M,
				MaxPoints: 999,
				Color:     models.BandColor_BLUE,
				Day:       2,
			},
			{
				Name:      "W",
				Sex:       models.BandSex_ALL,
				MaxPoints: 999,
				Color:     models.BandColor_BROWN,
				Day:       2,
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
				Confirmed: false,
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
			{
				MemberID:  member.ID,
				BandID:    bands[1].ID,
				Confirmed: false,
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
			{
				MemberID:  member.ID,
				BandID:    bands[2].ID,
				Confirmed: false,
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
			{
				MemberID:  member.ID,
				BandID:    bands[3].ID,
				Confirmed: false,
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
			{
				MemberID:  member.ID,
				BandID:    bands[4].ID,
				Confirmed: false,
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
		}
		require.NoError(t, env.db.Create(&entries).Error)

		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		data := map[string]interface{}{
			"BandIDs": []string{
				bands[0].ID.String(),
				bands[1].ID.String(),
				bands[2].ID.String(),
				bands[3].ID.String(),
				bands[4].ID.String(),
			},
			"SessionID": uuid.New(),
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var actual map[string]string
		require.NoError(t, json.NewDecoder(res.Body).Decode(&actual))
		require.NoError(t, err)
		require.Equal(t, http.StatusConflict, res.Code)
		require.Equal(t, limitThreeBandsPerDayReachedError.Error(), actual["error"])

		var updatedEntries []models.Entry
		require.NoError(t, env.db.Where(&models.Entry{MemberID: member.ID, Confirmed: true}).Order("created_at ASC").Find(&updatedEntries).Error)
		require.Len(t, updatedEntries, 0)
	})
	t.Run("LimitPerColorPerDayReached", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		bands := []models.Band{
			{
				Name:      "S",
				Sex:       models.BandSex_M,
				MaxPoints: 799,
				Color:     models.BandColor_PINK,
				Day:       1,
			},
			{
				Name:      "T",
				Sex:       models.BandSex_ALL,
				MaxPoints: 999,
				Color:     models.BandColor_PINK,
				Day:       1,
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
				Confirmed: false,
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
			{
				MemberID:  member.ID,
				BandID:    bands[1].ID,
				Confirmed: false,
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: sessionID,
			},
		}
		require.NoError(t, env.db.Create(&entries).Error)

		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		data := map[string]interface{}{
			"BandIDs": []string{
				bands[0].ID.String(),
				bands[1].ID.String(),
			},
			"SessionID": uuid.New(),
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var actual map[string]string
		require.NoError(t, json.NewDecoder(res.Body).Decode(&actual))
		require.NoError(t, err)
		require.Equal(t, http.StatusConflict, res.Code)
		require.Equal(t, limitSameColorPerDayReachedError.Error(), actual["error"])

		var updatedEntries []models.Entry
		require.NoError(t, env.db.Where(&models.Entry{MemberID: member.ID, Confirmed: true}).Order("created_at ASC").Find(&updatedEntries).Error)
		require.Len(t, updatedEntries, 0)
	})
	t.Run("NoMatchingSessionID", func(t *testing.T) {
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

		entries := []models.Entry{
			{
				MemberID:  member.ID,
				BandID:    bands[0].ID,
				Confirmed: false,
				CreatedAt: time.Now().Add(1 * time.Second),
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: uuid.New(),
			},
			{
				MemberID:  member.ID,
				BandID:    bands[1].ID,
				Confirmed: false,
				CreatedAt: time.Now().Add(2 * time.Second),
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: uuid.New(),
			},
		}
		require.NoError(t, env.db.Create(&entries).Error)

		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		data := map[string]interface{}{
			"BandIDs": []string{
				bands[0].ID.String(),
				bands[1].ID.String(),
			},
			"SessionID": uuid.New(),
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusConflict, res.Code)

		var updatedEntries []models.Entry
		require.NoError(t, env.db.Where(&models.Entry{MemberID: member.ID, Confirmed: true}).Order("created_at ASC").Find(&updatedEntries).Error)
		require.Len(t, updatedEntries, 0)
	})
	t.Run("OnlyOneMatchingSessionID", func(t *testing.T) {
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

		entries := []models.Entry{
			{
				MemberID:  member.ID,
				BandID:    bands[0].ID,
				Confirmed: false,
				CreatedAt: time.Now().Add(1 * time.Second),
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: uuid.New(),
			},
			{
				MemberID:  member.ID,
				BandID:    bands[1].ID,
				Confirmed: false,
				CreatedAt: time.Now().Add(2 * time.Second),
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: uuid.New(),
			},
		}
		require.NoError(t, env.db.Create(&entries).Error)

		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		data := map[string]interface{}{
			"BandIDs": []string{
				bands[0].ID.String(),
				bands[1].ID.String(),
			},
			"SessionID": entries[0].SessionID,
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusConflict, res.Code)

		var updatedEntries []models.Entry
		require.NoError(t, env.db.Where(&models.Entry{MemberID: member.ID, Confirmed: true}).Order("created_at ASC").Find(&updatedEntries).Error)
		require.Len(t, updatedEntries, 0)
	})
	t.Run("EntriesExpired", func(t *testing.T) {
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
				Confirmed: false,
				CreatedAt: time.Now().Add(1 * time.Second),
				ExpiresAt: time.Now().Add(-time.Hour),
				SessionID: sessionID,
			},
			{
				MemberID:  member.ID,
				BandID:    bands[1].ID,
				Confirmed: false,
				CreatedAt: time.Now().Add(2 * time.Second),
				ExpiresAt: time.Now().Add(-time.Hour),
				SessionID: sessionID,
			},
		}
		require.NoError(t, env.db.Create(&entries).Error)

		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		data := map[string]interface{}{
			"BandIDs": []string{
				bands[0].ID.String(),
				bands[1].ID.String(),
			},
			"SessionID": sessionID,
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusConflict, res.Code)

		var updatedEntries []models.Entry
		require.NoError(t, env.db.Where(&models.Entry{MemberID: member.ID, Confirmed: true}).Order("created_at ASC").Find(&updatedEntries).Error)
		require.Len(t, updatedEntries, 0)
	})
	t.Run("AllAlreadyConfirmed", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		bands := []models.Band{
			{
				Name:      "S",
				Sex:       models.BandSex_M,
				MaxPoints: 799,
				Color:     models.BandColor_GREEN,
				Day:       1,
			},
			{
				Name:      "T",
				Sex:       models.BandSex_M,
				MaxPoints: 999,
				Color:     models.BandColor_PINK,
				Day:       1,
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

		entries := []models.Entry{
			{
				MemberID:  member.ID,
				BandID:    bands[0].ID,
				Confirmed: true,
				CreatedAt: time.Now().Add(1 * time.Second),
				ExpiresAt: time.Now().Add(-time.Hour),
				SessionID: uuid.New(),
			},
			{
				MemberID:  member.ID,
				BandID:    bands[1].ID,
				Confirmed: true,
				CreatedAt: time.Now().Add(2 * time.Second),
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: uuid.New(),
			},
		}
		require.NoError(t, env.db.Create(&entries).Error)

		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		data := map[string]interface{}{
			"BandIDs": []string{
				bands[0].ID.String(),
				bands[1].ID.String(),
			},
			"SessionID": entries[0].SessionID,
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
		require.Equal(t, entries[0].ID, updatedEntries[0].ID)
		require.Equal(t, entries[0].BandID, updatedEntries[0].BandID)
		require.True(t, entries[0].CreatedAt.Equal(updatedEntries[0].CreatedAt))
		require.Equal(t, entries[0].MemberID, updatedEntries[0].MemberID)
		require.Equal(t, entries[0].SessionID, updatedEntries[0].SessionID)
		require.True(t, updatedEntries[0].Confirmed)
		require.Equal(t, entries[1].ID, updatedEntries[1].ID)
		require.Equal(t, entries[1].BandID, updatedEntries[1].BandID)
		require.True(t, entries[1].CreatedAt.Equal(updatedEntries[1].CreatedAt))
		require.Equal(t, entries[1].MemberID, updatedEntries[1].MemberID)
		require.Equal(t, entries[1].SessionID, updatedEntries[1].SessionID)
		require.True(t, updatedEntries[1].Confirmed)
	})
	t.Run("BandsNotFound", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		member := models.Member{
			FirstName: "John",
			LastName:  "Doe",
			Sex:       "M",
			PermitID:  "000000",
			Points:    700,
			UserID:    env.user.ID,
		}
		require.NoError(t, env.db.Create(&member).Error)

		band := models.Band{
			Name:      "S",
			Sex:       models.BandSex_M,
			MaxPoints: 799,
		}
		require.NoError(t, env.db.Create(&band).Error)

		url := fmt.Sprintf("/api/members/%s/set-entries", member.ID)
		invalidBandIDs := []string{uuid.NewString(), uuid.NewString()}
		data := map[string]interface{}{
			"BandIDs":   append(invalidBandIDs, band.ID.String()),
			"SessionID": uuid.NewString(),
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusNotFound, res.Code)
		var actual map[string]string
		require.NoError(t, json.NewDecoder(res.Body).Decode(&actual))
		require.Equal(t, fmt.Sprintf("bands %s not found", invalidBandIDs), actual["error"])
	})
	t.Run("BadUser", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		user := models.User{
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
		env.db.Create(&user)

		url := fmt.Sprintf("/api/members/%s/set-entries", user.Members[0].ID)
		data := map[string]interface{}{
			"BandIDs": []string{
				uuid.NewString(),
			},
			"SessionID": uuid.NewString(),
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusNotFound, res.Code)
		var actual map[string]string
		require.NoError(t, json.NewDecoder(res.Body).Decode(&actual))
		require.Equal(t, fmt.Sprintf("member %s not found", user.Members[0].ID), actual["error"])
	})
}
