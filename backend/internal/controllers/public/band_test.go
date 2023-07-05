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

func TestSetMemberBands(t *testing.T) {
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

		memberBands := []models.BandMember{
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

		url := fmt.Sprintf("/api/members/%s/set-bands", member.ID)
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

		var updated models.Member
		err = json.NewDecoder(res.Body).Decode(&updated)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.Code)
		require.Len(t, updated.Bands, 2)
		require.Equal(t, bands[0], *updated.Bands[0])
		require.Equal(t, bands[2], *updated.Bands[1])
	})
}
