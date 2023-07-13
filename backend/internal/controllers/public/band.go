package public

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/SuperPingPong/tournoi/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/getsentry/sentry-go"
)

func (api *API) ListBands(ctx *gin.Context) {
	var bands []models.Band
	var filteredBands *gorm.DB

	day, err := strconv.Atoi(ctx.Query("day"))
	if err != nil {
		filteredBands = api.db
	} else {
		filteredBands = api.db.Where("day = ?", day)
	}

	err = filteredBands.Find(&bands).Error
	if err != nil {
		sentry.CaptureException(fmt.Errorf("failed to list bands: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list bands: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"bands": bands})
}
