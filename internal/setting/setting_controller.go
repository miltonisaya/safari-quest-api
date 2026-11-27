package setting

import (
	"errors"
	"net/http"

	"safari-quest-api/internal/user"
	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct{}

// Index godoc
//
//	@Summary		List system settings
//	@Description	Every setting with its description, allowed range, default and current value.
//	@Tags			System settings
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.CustomApiResponse{data=[]setting.Response}
//	@Router			/system-settings [get]
func (Controller) Index(c *gin.Context) {
	items, err := GetAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Could not load settings")
		return
	}
	response.Success(c, http.StatusOK, "System settings", items)
}

// Update godoc
//
//	@Summary		Change system settings
//	@Description	Saves the given values after checking each against its allowed range; nothing is saved if one is invalid.
//	@Tags			System settings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		setting.UpdateInput	true	"Values by key"
//	@Success		200		{object}	response.CustomApiResponse{data=[]setting.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Router			/system-settings [put]
func (Controller) Update(c *gin.Context) {
	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Send the settings to change", gin.H{"error": err.Error()})
		return
	}
	var userID uint
	if v, ok := c.Get("userUUID"); ok {
		if id, ok := v.(uuid.UUID); ok {
			if u, err := user.FindByUUID(id); err == nil {
				userID = u.ID
			}
		}
	}
	items, err := Update(input, userID)
	if err != nil {
		if errors.Is(err, ErrUnknownSetting) || errors.Is(err, ErrOutOfRange) {
			response.Fail(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Could not save settings")
		return
	}
	response.Success(c, http.StatusOK, "Settings saved", items)
}
