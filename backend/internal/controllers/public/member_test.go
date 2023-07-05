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
	type listMembersResponse struct {
		Members []models.Member
	}
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
				FirstName: "John",
				LastName:  "Doe",
				Sex:       "M",
				PermitID:  "000000",
				UserID:    env.user.ID,
			},
			{
				FirstName: "Jane",
				LastName:  "Doe",
				Sex:       "F",
				PermitID:  "000001",
				UserID:    env.user.ID,
			},
		}
		env.db.Create(&members)

		res := performRequest("GET", "/api/members", nil, map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		require.Equal(t, http.StatusOK, res.Code)
		var got listMembersResponse
		err := json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.Len(t, got.Members, 2)
		require.Equal(t, members[0], got.Members[0])
		require.Equal(t, members[1], got.Members[1])
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
		var got listMembersResponse
		err := json.NewDecoder(res.Body).Decode(&got)
		require.NoError(t, err)
		require.Len(t, got.Members, 0)
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
		require.Equal(t, "invalid input: Key: 'CreateMemberInput.MemberID' Error:Field validation for 'MemberID' failed on the 'required' tag", actual["error"])
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
		data := map[string]string{"MemberID": "000000"}
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
		data := map[string]string{"MemberID": "000000"}
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
