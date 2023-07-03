package public

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckAuth(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		data := map[string]string{"email": env.user.Email}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", "/api/check-auth", bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var response struct {
			Valid bool
		}
		require.Equal(t, http.StatusOK, res.Code)
		err = json.NewDecoder(res.Body).Decode(&response)
		require.NoError(t, err)
		require.True(t, response.Valid)
	})
	t.Run("NoAuth", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		data := map[string]string{"email": env.user.Email}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", "/api/check-auth", bytes.NewBuffer(body), map[string]string{}, env.api.router)

		require.Equal(t, http.StatusUnauthorized, res.Code)
	})
	t.Run("WrongEmail", func(t *testing.T) {
		env := getTestEnv(t)
		defer env.teardown()

		data := map[string]string{"email": "hdupont@example.com"}
		body, err := json.Marshal(data)
		require.NoError(t, err)

		res := performRequest("POST", "/api/check-auth", bytes.NewBuffer(body), map[string]string{
			"Authorization": "Bearer " + env.jwt,
		}, env.api.router)

		var response struct {
			Valid bool
		}
		require.Equal(t, http.StatusOK, res.Code)
		err = json.NewDecoder(res.Body).Decode(&response)
		require.NoError(t, err)
		require.False(t, response.Valid)
	})
}
