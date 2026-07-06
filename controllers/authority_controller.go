package controllers

import (
	"net/http"

	"safari-quest-api/pkg/response"
	"safari-quest-api/services"

	"github.com/gin-gonic/gin"
)

type AuthorityController struct{}

// Index godoc
//
//	@Summary		List authorities
//	@Description	Return all authorities
//	@Tags			Authorities
//	@Produce		json
//	@Success		200	{object}	response.CustomApiResponse{data=[]services.AuthorityListResponse}
//	@Failure		500	{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/authorities [get]
func (controller AuthorityController) Index(c *gin.Context) {
	authorities, err := services.AuthorityGetAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Authorities retrieved successfully", authorities)
}
