package public

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/SuperPingPong/tournoi/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListMembers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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

		members := []models.Member{
			{
				FirstName:  "John",
				LastName:   "Doe",
				Sex:        "M",
				PermitID:   "000000",
				Points:     100.0,
				Category:   "V2",
				ClubName:   "Jane Club",
				PermitType: "T",
				UserID:     env.user.ID,
				CreatedAt:  time.Now().Add(1 * time.Minute),
			},
			{
				FirstName:  "Jane",
				LastName:   "Doe",
				Sex:        "F",
				PermitID:   "000001",
				Points:     130.0,
				Category:   "B1",
				ClubName:   "Jane Club",
				PermitType: "P",
				UserID:     env.user.ID,
				CreatedAt:  time.Now(),
			},
		}
		env.db.Create(&members)

		bands := []models.Band{
			{
				Name:      "A",
				Day:       1,
				Color:     models.BandColor_BLUE,
				CreatedAt: time.Now().Add(-3 * time.Second),
			},
			{
				Name:      "B",
				Day:       2,
				Color:     models.BandColor_BROWN,
				CreatedAt: time.Now().Add(-2 * time.Second),
			},
			{
				Name:      "C",
				Day:       2,
				Color:     models.BandColor_GREEN,
				CreatedAt: time.Now().Add(-1 * time.Second),
			},
		}
		env.db.Create(&bands)

		entries := []models.Entry{
			{
				MemberID:  members[0].ID,
				BandID:    bands[0].ID,
				Confirmed: true,
			},
			{
				MemberID:  members[0].ID,
				BandID:    bands[1].ID,
				Confirmed: true,
			},
			{
				MemberID:  members[0].ID,
				BandID:    bands[2].ID,
				Confirmed: false,
			},
		}
		env.db.Create(&entries)

		res := performRequest("GET", "/api/members", nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		var got ListMembersMembers
		err := json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.Len(t, got.Members, 2)
		require.Equal(t, 2, got.Total)

		require.Equal(t, members[0].ID, got.Members[0].ID)
		require.Equal(t, members[0].FirstName, got.Members[0].FirstName)
		require.Equal(t, members[0].LastName, got.Members[0].LastName)
		require.Equal(t, members[0].PermitID, got.Members[0].PermitID)
		require.Equal(t, members[0].Sex, got.Members[0].Sex)
		require.Equal(t, members[0].Points, got.Members[0].Points)
		require.Equal(t, members[0].Category, got.Members[0].Category)
		require.Equal(t, members[0].ClubName, got.Members[0].ClubName)
		require.Equal(t, members[0].PermitType, got.Members[0].PermitType)

		// Only two entries are expected since we don't want to show non-confirmed entries
		require.Len(t, got.Members[0].Entries, 2)
		require.Equal(t, bands[0].ID, got.Members[0].Entries[0].BandID)
		require.Equal(t, bands[0].Name, got.Members[0].Entries[0].BandName)
		require.Equal(t, entries[0].CreatedAt, got.Members[0].Entries[0].CreatedAt)
		require.Equal(t, bands[1].ID, got.Members[0].Entries[1].BandID)
		require.Equal(t, bands[1].Name, got.Members[0].Entries[1].BandName)
		require.Equal(t, entries[1].CreatedAt, got.Members[0].Entries[1].CreatedAt)

		require.Equal(t, members[1].ID, got.Members[1].ID)
		require.Equal(t, members[1].FirstName, got.Members[1].FirstName)
		require.Equal(t, members[1].LastName, got.Members[1].LastName)
		require.Equal(t, members[1].PermitID, got.Members[1].PermitID)
		require.Equal(t, members[1].Sex, got.Members[1].Sex)
		require.Equal(t, members[1].Points, got.Members[1].Points)
		require.Equal(t, members[1].Category, got.Members[1].Category)
		require.Equal(t, members[1].ClubName, got.Members[1].ClubName)
		require.Equal(t, members[1].PermitType, got.Members[1].PermitType)
	})
	t.Run("SuccessSearch", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		members := []models.Member{
			{
				FirstName:  "John",
				LastName:   "Doe",
				Sex:        "M",
				PermitID:   "000000",
				Points:     600.0,
				Category:   "V2",
				ClubName:   "Jane Club",
				PermitType: "T",
				UserID:     env.user.ID,
				CreatedAt:  time.Now().Add(-3 * time.Second),
			},
			{
				FirstName:  "Jane",
				LastName:   "Doe",
				Sex:        "F",
				PermitID:   "000001",
				Points:     700.0,
				Category:   "B1",
				ClubName:   "Jane Club",
				PermitType: "P",
				UserID:     env.user.ID,
				CreatedAt:  time.Now().Add(-2 * time.Second),
			},
			{
				FirstName:  "Hervé",
				LastName:   "Dupont",
				Sex:        "M",
				PermitID:   "000003",
				ClubName:   "Club du Pont Hervé",
				Points:     505.0,
				Category:   "V3",
				PermitType: "P",
				UserID:     env.user.ID,
				CreatedAt:  time.Now().Add(-1 * time.Second),
			},
		}
		env.db.Create(&members)

		url := "/api/members?search=%s&order_by=created_at_asc"

		// search=doe should return John and Jane
		res := performRequest("GET", fmt.Sprintf(url, "doe"), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var got ListMembersMembers
		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Members, 2)
		require.Equal(t, 2, got.Total)
		require.Equal(t, members[0].ID, got.Members[0].ID)
		require.Equal(t, members[1].ID, got.Members[1].ID)

		// search=CLUB should return all of them
		res = performRequest("GET", fmt.Sprintf(url, "CLUB"), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Members, 3)
		require.Equal(t, 3, got.Total)
		require.Equal(t, members[0].ID, got.Members[0].ID)
		require.Equal(t, members[1].ID, got.Members[1].ID)
		require.Equal(t, members[2].ID, got.Members[2].ID)

		// search=Test should return none of them since they belong to the user test@example.com but non-admin user
		// should not be able to search on the email field
		res = performRequest("GET", fmt.Sprintf(url, "Test"), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Members, 0)
		require.Equal(t, 0, got.Total)

		// search=george should return none of them
		res = performRequest("GET", fmt.Sprintf(url, "george"), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Members, 0)
		require.Equal(t, 0, got.Total)
	})
	t.Run("SuccessFilterByPermitID", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		members := []models.Member{
			{
				FirstName:  "John",
				LastName:   "Doe",
				Sex:        "M",
				PermitID:   "000001",
				Points:     600.0,
				Category:   "V2",
				ClubName:   "Jane Club",
				PermitType: "T",
				UserID:     env.user.ID,
			},
			{
				FirstName:  "Jane",
				LastName:   "Doe",
				Sex:        "F",
				PermitID:   "000002",
				Points:     700.0,
				Category:   "B1",
				ClubName:   "Jane Club",
				PermitType: "P",
				UserID:     env.user.ID,
			},
		}
		env.db.Create(&members)

		url := "/api/members?permit_id=%s"

		// search=doe should return John and Jane
		res := performRequest("GET", fmt.Sprintf(url, members[0].PermitID), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var got ListMembersMembers
		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Members, 1)
		require.Equal(t, 1, got.Total)
		require.Equal(t, members[0].ID, got.Members[0].ID)
	})
	t.Run("NoMemberForCurrentUser", func(t *testing.T) {
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

		res := performRequest("GET", "/api/members", nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		var got ListMembersMembers
		err := json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.Equal(t, ListMembersMembers{
			Members: []ListMembersMember{},
			Total:   0,
		}, got)
	})
	t.Run("SuccessAdmin", func(t *testing.T) {
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
					CreatedAt: time.Now().Add(-1 * time.Second),
				},
			},
		}
		env.db.Create(&user)

		members := []models.Member{
			{
				FirstName:  "John",
				LastName:   "Doe",
				Sex:        "M",
				PermitID:   "000000",
				Points:     100.0,
				Category:   "V2",
				ClubName:   "Jane Club",
				PermitType: "T",
				UserID:     env.user.ID,
				CreatedAt:  time.Now().Add(-2 * time.Second),
			},
			{
				FirstName:  "Jane",
				LastName:   "Doe",
				Sex:        "F",
				PermitID:   "000001",
				Points:     130.0,
				Category:   "B1",
				ClubName:   "Jane Club",
				PermitType: "P",
				UserID:     env.user.ID,
				CreatedAt:  time.Now().Add(-3 * time.Second),
			},
		}
		env.db.Create(&members)

		res := performRequest("GET", "/api/members", nil, map[string]string{
			"Authorization": "Bearer " + env.adminJWT,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		var got ListMembersMembers
		err := json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.Len(t, got.Members, 3)
		require.Equal(t, 3, got.Total)

		require.Equal(t, user.Members[0].ID, got.Members[0].ID)
		require.Equal(t, user.Members[0].FirstName, got.Members[0].FirstName)
		require.Equal(t, user.Members[0].LastName, got.Members[0].LastName)
		require.Equal(t, user.Members[0].PermitID, got.Members[0].PermitID)
		require.Equal(t, user.Members[0].Sex, got.Members[0].Sex)
		require.Equal(t, user.Members[0].Points, got.Members[0].Points)
		require.Equal(t, user.Members[0].Category, got.Members[0].Category)
		require.Equal(t, user.Members[0].ClubName, got.Members[0].ClubName)
		require.Equal(t, user.Members[0].PermitType, got.Members[0].PermitType)

		require.Equal(t, members[0].ID, got.Members[1].ID)
		require.Equal(t, members[0].FirstName, got.Members[1].FirstName)
		require.Equal(t, members[0].LastName, got.Members[1].LastName)
		require.Equal(t, members[0].PermitID, got.Members[1].PermitID)
		require.Equal(t, members[0].Sex, got.Members[1].Sex)
		require.Equal(t, members[0].Points, got.Members[1].Points)
		require.Equal(t, members[0].Category, got.Members[1].Category)
		require.Equal(t, members[0].ClubName, got.Members[1].ClubName)
		require.Equal(t, members[0].PermitType, got.Members[1].PermitType)

		require.Equal(t, members[1].ID, got.Members[2].ID)
		require.Equal(t, members[1].FirstName, got.Members[2].FirstName)
		require.Equal(t, members[1].LastName, got.Members[2].LastName)
		require.Equal(t, members[1].PermitID, got.Members[2].PermitID)
		require.Equal(t, members[1].Sex, got.Members[2].Sex)
		require.Equal(t, members[1].Points, got.Members[2].Points)
		require.Equal(t, members[1].Category, got.Members[2].Category)
		require.Equal(t, members[1].ClubName, got.Members[2].ClubName)
		require.Equal(t, members[1].PermitType, got.Members[2].PermitType)
	})
	t.Run("SuccessSearchAdmin", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		members := []models.Member{
			{
				FirstName:  "John",
				LastName:   "Doe",
				Sex:        "M",
				PermitID:   "000000",
				Points:     600.0,
				Category:   "V2",
				ClubName:   "Jane Club",
				PermitType: "T",
				UserID:     env.user.ID,
				CreatedAt:  time.Now().Add(-3 * time.Second),
			},
			{
				FirstName:  "Jane",
				LastName:   "Doe",
				Sex:        "F",
				PermitID:   "000001",
				Points:     700.0,
				Category:   "B1",
				ClubName:   "Jane Club",
				PermitType: "P",
				UserID:     env.user.ID,
				CreatedAt:  time.Now().Add(-2 * time.Second),
			},
			{
				FirstName:  "Hervé",
				LastName:   "Dupont",
				Sex:        "M",
				PermitID:   "000003",
				ClubName:   "Club du Pont Hervé",
				Points:     505.0,
				Category:   "V3",
				PermitType: "P",
				UserID:     env.user.ID,
				CreatedAt:  time.Now().Add(-1 * time.Second),
			},
		}
		env.db.Create(&members)

		url := "/api/members?search=%s&order_by=created_at_asc"

		// search=doe should return John and Jane
		res := performRequest("GET", fmt.Sprintf(url, "doe"), nil, map[string]string{
			"Authorization": "Bearer " + env.adminJWT,
		}, env.api.router)

		var got ListMembersMembers
		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Members, 2)
		require.Equal(t, 2, got.Total)
		require.Equal(t, members[0].ID, got.Members[0].ID)
		require.Equal(t, members[1].ID, got.Members[1].ID)

		// search=CLUB should return all of them
		res = performRequest("GET", fmt.Sprintf(url, "CLUB"), nil, map[string]string{
			"Authorization": "Bearer " + env.adminJWT,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Members, 3)
		require.Equal(t, 3, got.Total)
		require.Equal(t, members[0].ID, got.Members[0].ID)
		require.Equal(t, members[1].ID, got.Members[1].ID)
		require.Equal(t, members[2].ID, got.Members[2].ID)

		// search=Test should return all of them since they all belong to the user test@example.com
		res = performRequest("GET", fmt.Sprintf(url, "Test"), nil, map[string]string{
			"Authorization": "Bearer " + env.adminJWT,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Members, 3)
		require.Equal(t, 3, got.Total)
		require.Equal(t, members[0].ID, got.Members[0].ID)
		require.Equal(t, members[1].ID, got.Members[1].ID)
		require.Equal(t, members[2].ID, got.Members[2].ID)

		// search=george should return none of them
		res = performRequest("GET", fmt.Sprintf(url, "george"), nil, map[string]string{
			"Authorization": "Bearer " + env.adminJWT,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Members, 0)
		require.Equal(t, 0, got.Total)
	})
}

func TestCreateMember(t *testing.T) {
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

		data := map[string]string{"permitID": permitID}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", "/api/members", bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var created models.Member
		require.Equal(t, http.StatusCreated, res.Code)
		err = json.NewDecoder(res.Body).Decode(&created)
		require.NoError(t, err)
		require.NotEqual(t, created.ID, uuid.Nil)
		require.Equal(t, firstName, created.FirstName)
		require.Equal(t, lastName, created.LastName)
		require.Equal(t, permitID, created.PermitID)
		require.Equal(t, sex, created.Sex)
		require.Equal(t, point, created.Points)
		require.Equal(t, category, created.Category)
		require.Equal(t, clubName, created.ClubName)
		require.Equal(t, permitType, created.PermitType)
		require.Equal(t, env.user.ID, created.UserID)
		require.True(t, created.CreatedAt.After(time.Time{}))
		require.Equal(t, created.CreatedAt, created.UpdatedAt)
		require.False(t, created.DeletedAt.Valid)
	})
	t.Run("EmptyInput", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		data := map[string]string{}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", "/api/members", bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var actual map[string]string
		require.Equal(t, http.StatusBadRequest, res.Code)
		err = json.NewDecoder(res.Body).Decode(&actual)
		require.NoError(t, err)
		require.Equal(t, "invalid input: Key: 'CreateMemberInput.PermitID' Error:Field validation for 'PermitID' failed on the 'required' tag", actual["error"])
	})
	t.Run("MissingInput", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		res := performRequest("POST", "/api/members", nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var actual map[string]string
		err := json.NewDecoder(res.Body).Decode(&actual)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "invalid input: EOF", actual["error"])
	})
}
