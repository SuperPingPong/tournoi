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
			},
		}
		env.db.Create(&members)

		bands := []models.Band{
			{
				Name: "A",
				Day:  1,
			},
			{
				Name: "B",
				Day:  2,
			},
		}
		env.db.Create(&bands)

		entries := []models.Entry{
			{
				MemberID: members[0].ID,
				BandID:   bands[0].ID,
			},
			{
				MemberID: members[0].ID,
				BandID:   bands[1].ID,
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
			},
		}
		env.db.Create(&members)

		url := "/api/members?search=%s"

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

		// search=Test should return all of them since they all belong to the user test@example.com
		res = performRequest("GET", fmt.Sprintf(url, "Test"), nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
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
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
		require.Len(t, got.Members, 0)
		require.Equal(t, 0, got.Total)
	})
	t.Run("NoMemberForCurrentUser", func(t *testing.T) {
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

func TestUpdateMember(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		created := &models.Member{
			FirstName: "John",
			LastName:  "Doe",
			Sex:       "M",
			PermitID:  "000001",
			UserID:    env.user.ID,
		}
		env.db.Create(created)

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

		url := fmt.Sprintf("/api/members/%s", created.ID)
		data := map[string]string{
			"PermitID": permitID,
		}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("PATCH", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var updated models.Member
		err = json.NewDecoder(res.Body).Decode(&updated)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, created.ID, updated.ID)
		require.Equal(t, created.CreatedAt, updated.CreatedAt)
		require.True(t, created.UpdatedAt.Before(updated.UpdatedAt))
		require.Equal(t, firstName, updated.FirstName)
		require.Equal(t, lastName, updated.LastName)
		require.Equal(t, permitID, updated.PermitID)
		require.Equal(t, sex, updated.Sex)
		require.Equal(t, category, updated.Category)
		require.Equal(t, clubName, updated.ClubName)
		require.Equal(t, permitType, updated.PermitType)
		require.Equal(t, env.user.ID, updated.UserID)
	})
	t.Run("EmptyInput", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		url := "/api/members/000001"
		data := map[string]string{}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		res := performRequest("PATCH", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var updated models.Member
		err = json.NewDecoder(res.Body).Decode(&updated)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.Code)
	})
	t.Run("MissingInput", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		created := &models.Member{
			FirstName: "John",
			LastName:  "Doe",
			Sex:       "M",
			PermitID:  "123456",
		}
		env.db.Create(created)

		url := fmt.Sprintf("/api/members/%s", created.ID)
		res := performRequest("PATCH", url, nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var actual map[string]string
		require.Equal(t, http.StatusBadRequest, res.Code)
		err := json.NewDecoder(res.Body).Decode(&actual)
		require.NoError(t, err)
		require.Equal(t, "invalid input: EOF", actual["error"])
	})
	t.Run("InvalidMemberID", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		url := "/api/members/foo"
		res := performRequest("PATCH", url, nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var actual map[string]string
		err := json.NewDecoder(res.Body).Decode(&actual)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "invalid member id: foo", actual["error"])
	})
	t.Run("WrongUserID", func(t *testing.T) {
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

		url := "/api/members/" + user.Members[0].ID.String()
		data := map[string]string{"PermitID": "000000"}
		body, err := json.Marshal(data)
		require.NoError(t, err)
		res := performRequest("PATCH", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var actual map[string]string
		err = json.NewDecoder(res.Body).Decode(&actual)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, res.Code)
		require.Equal(t, fmt.Sprintf("member %s not found", user.Members[0].ID.String()), actual["error"])
	})
	t.Run("NotFound", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		memberID := uuid.NewString()
		url := "/api/members/" + memberID
		data := map[string]string{"PermitID": "000000"}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("PATCH", url, bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var actual map[string]string
		err = json.NewDecoder(res.Body).Decode(&actual)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, res.Code)
		require.Equal(t, fmt.Sprintf("member %s not found", memberID), actual["error"])
	})
}
