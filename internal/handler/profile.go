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

type ProfileRequest struct {
	Username  string  `json:"username" binding:"required"`
	Bio       *string `json:"bio" extensions:"x-nullable"`
	AvatarURL *string `json:"avatarUrl" extensions:"x-nullable"`
}

type ProfileResponse struct {
	UserID    string    `json:"userId" format:"uuid"`
	Username  string    `json:"username"`
	Bio       *string   `json:"bio" extensions:"x-nullable"`
	AvatarURL *string   `json:"avatarUrl" extensions:"x-nullable"`
	CreatedAt time.Time `json:"createdAt" format:"date-time"`
	UpdatedAt time.Time `json:"updatedAt" format:"date-time"`
}

// GetProfile godoc
// @Summary      Получение своего профиля
// @Description  Возвращает профиль пользователя из заголовка X-User-ID.
// @Tags         profile
// @Produce      json
// @Security     XUserID
// @Success      200  {object}  ProfileResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /v1/profile [get]
func (h profileHandler) get(c *gin.Context) {
	profile, err := h.service.Get(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, profileDTO(profile))
}

// UpsertProfile godoc
// @Summary      Создание или обновление своего профиля
// @Description  Полностью заменяет редактируемые поля профиля пользователя из заголовка X-User-ID.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     XUserID
// @Param        profile  body      ProfileRequest  true  "Поля профиля"
// @Success      200      {object}  ProfileResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      413      {object}  ErrorResponse
// @Failure      415      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /v1/profile [put]
func (h profileHandler) upsert(c *gin.Context) {
	var request ProfileRequest
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

func profileDTO(profile domain.Profile) ProfileResponse {
	return ProfileResponse{
		UserID: profile.UserID, Username: profile.Username, Bio: profile.Bio,
		AvatarURL: profile.AvatarURL, CreatedAt: profile.CreatedAt, UpdatedAt: profile.UpdatedAt,
	}
}
