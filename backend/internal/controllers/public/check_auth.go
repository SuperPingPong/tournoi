package public

import (
	"fmt"
	"net/http"

	"github.com/SuperPingPong/tournoi/internal/auth"
	"github.com/SuperPingPong/tournoi/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type CheckAuthInput struct {
	Email string `binding:"required,email"`
}

func (api *API) CheckAuth(ctx *gin.Context) {
	user, ok := ctx.Get(auth.IdentityKey)
	if !ok {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to extract current user"))
		return
	}

	var input CheckAuthInput
	err := ctx.ShouldBindBodyWith(&input, binding.JSON)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	var valid bool
	if user.(*models.User).Email == input.Email {
		valid = true
	}
	ctx.JSON(http.StatusOK, gin.H{"valid": valid})
}
