package handler

import (
	"net/http"
	"strings"

	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/VsemoGosi-Adventure/personal/internal/service"
	"github.com/gin-gonic/gin"
)

type authHandler struct{ service service.AuthService }

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Создает учетные данные и базовый профиль, возвращает пару токенов.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      registerRequest    true  "Данные регистрации"
// @Success      201      {object}  domain.AuthTokens
// @Failure      400      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /v1/auth/register [post]
func (h authHandler) register(c *gin.Context) {
	var request registerRequest
	if !decodeJSON(c, &request) {
		return
	}
	tokens, err := h.service.Register(c.Request.Context(), domain.RegisterInput{
		Email: request.Email, Password: request.Password, Username: request.Username,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, tokens)
}

// Login godoc
// @Summary      Вход пользователя
// @Description  Проверяет пароль и возвращает новую пару токенов.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      loginRequest       true  "Учетные данные"
// @Success      200      {object}  domain.AuthTokens
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /v1/auth/login [post]
func (h authHandler) login(c *gin.Context) {
	var request loginRequest
	if !decodeJSON(c, &request) {
		return
	}
	tokens, err := h.service.Login(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

// Refresh godoc
// @Summary      Обновление пары токенов
// @Description  Однократно заменяет refresh токен и возвращает новую пару.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      refreshRequest     true  "Refresh токен"
// @Success      200      {object}  domain.AuthTokens
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /v1/auth/refresh [post]
func (h authHandler) refresh(c *gin.Context) {
	var request refreshRequest
	if !decodeJSON(c, &request) {
		return
	}
	tokens, err := h.service.Refresh(c.Request.Context(), request.RefreshToken)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

// Logout godoc
// @Summary      Выход пользователя
// @Description  Инвалидирует переданный refresh токен.
// @Tags         auth
// @Accept       json
// @Param        request  body      refreshRequest     true  "Refresh токен"
// @Success      204      "Сессия завершена"
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /v1/auth/logout [post]
func (h authHandler) logout(c *gin.Context) {
	var request refreshRequest
	if !decodeJSON(c, &request) {
		return
	}
	if err := h.service.Logout(c.Request.Context(), request.RefreshToken); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Verify godoc
// @Summary      Проверка access токена для Traefik ForwardAuth
// @Description  Проверяет Bearer JWT и возвращает ID пользователя в заголовке X-User-ID.
// @Tags         auth
// @Security     BearerAuth
// @Success      200  "Токен действителен"
// @Header       200  {string}  X-User-ID  "UUID пользователя"
// @Failure      401  {object}  ErrorResponse
// @Router       /v1/auth/verify [get]
func (h authHandler) verify(c *gin.Context) {
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		writeError(c, http.StatusUnauthorized, "unauthorized", "invalid credentials or token")
		return
	}
	userID, err := h.service.Verify(parts[1])
	if err != nil {
		writeError(c, http.StatusUnauthorized, "unauthorized", "invalid credentials or token")
		return
	}
	c.Header("X-User-ID", userID)
	c.Status(http.StatusOK)
}
