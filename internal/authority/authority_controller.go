package authority

import (
	"net/http"

	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type Controller struct{}

// Index godoc
//
//	@Summary		List authorities
//	@Description	Return all authorities
//	@Tags			Authorities
//	@Produce		json
//	@Success		200	{object}	response.CustomApiResponse{data=[]ListResponse}
//	@Failure		500	{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/authorities [get]
func (controller Controller) Index(c *gin.Context) {
	authorities, err := GetAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Authorities retrieved successfully", authorities)
}
