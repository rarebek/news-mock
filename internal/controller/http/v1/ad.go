package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/k0kubun/pp"
	"tarkib.uz/internal/entity"
	"tarkib.uz/internal/usecase"
	"tarkib.uz/pkg/logger"
	tokens "tarkib.uz/pkg/token"
)

type adRoutes struct {
	t usecase.AdUseCase
	l logger.Interface
}

type Claims struct {
	Role string `json:"role"`
	jwt.StandardClaims
}

func newAdRoutes(handler *gin.RouterGroup, t usecase.AdUseCase, l logger.Interface) {
	r := &adRoutes{t, l}
	h := handler.Group("/ads")
	{
		h.POST("/", r.createAd)
		h.DELETE("/:id", r.deleteAd)
		h.PUT("/", r.updateAd)
		h.GET("/:id", r.getAd)
		h.GET("/all", r.getAllAds)
	}
}

// @Summary     Create a new ad
// @Description Create a new ad with the given details
// @Tags        ads
// @Accept      json
// @Produce     json
// @Param       ad body entity.CreateAdRequest true "Ad details"
// @Success     201 {object} entity.Ad
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /ads/ [post]
func (r *adRoutes) createAd(c *gin.Context) {
	var ad entity.CreateAdRequest
	if err := c.ShouldBindJSON(&ad); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, "Invalid request body", false)
		return
	}
	id := uuid.NewString()
	err := r.t.CreateAd(c.Request.Context(), &entity.Ad{
		ID:        id,
		Link:      ad.Link,
		ImageURL:  ad.ImageURL,
		CreatedAt: time.Now(),
	})
	if err != nil {
		if err.Error() == "an ad already exists" {
			r.l.Error(err)
			errorResponse(c, http.StatusConflict, "ad is already exists", false)
			return
		}
		r.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to create ad", false)
		return
	}

	var response entity.Ad
	response.ID = id
	response.ImageURL = ad.ImageURL
	response.Link = ad.Link
	response.CreatedAt = time.Now()

	c.JSON(http.StatusCreated, response)
}

// @Summary     Delete an ad
// @Description Delete an ad
// @Tags        ads
// @Produce     json
// @Param       id path string true "ID of the ads to delete"
// @Success     204
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /ads/{id} [delete]
func (r *adRoutes) deleteAd(c *gin.Context) {
	id := c.Param("id")

	if err := r.t.DeleteAd(c.Request.Context(), id); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to delete ad", false)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary		Update Ad
// @Description Edits ad by ID
// @Tags        ads
// @Produce     json
// @Param       ad body entity.CreateAdRequest true "Ad details"
// @Success     204
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /ads [put]
func (r *adRoutes) updateAd(c *gin.Context) {
	var ad entity.Ad

	if err := c.ShouldBindJSON(&ad); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusBadRequest, "Request body not matched", false)
		return
	}
	if err := r.t.UpdateAd(c.Request.Context(), &ad); err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to update ad", false)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary		Gets ad details
// @Description returns ads
// @Tags        ads
// @Produce     json
// @Param       id path string true "ID of the ads to get"
// @Success     200
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Security    BearerAuth
// @Router      /ads/{id} [get]
func (r *adRoutes) getAd(c *gin.Context) {
	id := c.Param("id")
	tokenStr := c.Request.Header.Get("Authorization")
	fmt.Println(tokenStr)
	if tokenStr == "" {
		ad, err := r.t.GetAd(c.Request.Context(), &entity.GetAdRequest{
			IsAdmin: false,
			ID:      id,
		})

		if err != nil {
			r.l.Error(err)
			errorResponse(c, http.StatusInternalServerError, "Failed to get ad"+err.Error(), false)
			return
		}

		c.JSON(http.StatusOK, ad)
	} else {
		pp.Println(tokenStr)
		jwt := tokens.JWTHandler{
			SigninKey: "dfhdghkglioe",
			Token:     tokenStr,
		}

		claims, err := jwt.ExtractClaims()
		if err != nil {
			r.l.Error(err)
			errorResponse(c, http.StatusInternalServerError, "Failed to get aaad"+err.Error(), false)
			return
		}

		if claims["role"] == "super-admin" {
			ad, err := r.t.GetAd(c.Request.Context(), &entity.GetAdRequest{
				IsAdmin: true,
				ID:      id,
			})
			if err != nil {
				r.l.Error(err)
				errorResponse(c, http.StatusInternalServerError, "Failed to get ad", false)
				return
			}

			c.JSON(http.StatusOK, ad)
		}
	}

}

// @Summary     Get all ads
// @Description Get all ads with view count for admins
// @Tags        ads
// @Produce     json
// @Success     200 {array} entity.Ad
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /ads/all [get]
func (r *adRoutes) getAllAds(c *gin.Context) {
	ads, err := r.t.GetAllAds(c.Request.Context())
	if err != nil {
		r.l.Error(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to get all ads: "+err.Error(), false)
		return
	}

	c.JSON(http.StatusOK, ads)
}
