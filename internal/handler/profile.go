package handler

import (
	"net/http"
	"time"

	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/VsemoGosi-Adventure/personal/internal/middleware"
	"github.com/VsemoGosi-Adventure/personal/internal/service"
	"github.com/gin-gonic/gin"
)

type profileHandler struct {
	service service.ProfileService
}

type profileRequest struct {
	Username  string  `json:"username"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatarUrl"`
}

type profileResponse struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Bio       *string   `json:"bio"`
	AvatarURL *string   `json:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (h profileHandler) get(c *gin.Context) {
	profile, err := h.service.Get(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, profileDTO(profile))
}

func (h profileHandler) upsert(c *gin.Context) {
	var request profileRequest
	if !decodeJSON(c, &request) {
		return
	}
	profile, err := h.service.Upsert(c.Request.Context(), middleware.UserID(c), domain.ProfileUpdate{
		Username: request.Username, Bio: request.Bio, AvatarURL: request.AvatarURL,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, profileDTO(profile))
}

func profileDTO(profile domain.Profile) profileResponse {
	return profileResponse{
		UserID: profile.UserID, Username: profile.Username, Bio: profile.Bio,
		AvatarURL: profile.AvatarURL, CreatedAt: profile.CreatedAt, UpdatedAt: profile.UpdatedAt,
	}
}
