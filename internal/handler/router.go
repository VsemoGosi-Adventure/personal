package handler

import (
	"net/http"

	"github.com/VsemoGosi-Adventure/personal/internal/middleware"
	"github.com/VsemoGosi-Adventure/personal/internal/service"
	"github.com/gin-gonic/gin"
)

// NewRouter exposes the profile and interest services over HTTP.
func NewRouter(profiles service.ProfileService, interests service.InterestService) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/healthz", health)

	api := router.Group("/v1", middleware.UserIDMiddleware())
	profile := profileHandler{service: profiles}
	interest := interestHandler{service: interests}
	api.GET("/profile", profile.get)
	api.PUT("/profile", profile.upsert)
	api.GET("/interests", interest.list)
	api.GET("/profile/interests", interest.listForUser)
	api.PUT("/profile/interests/:interestID", interest.add)
	api.DELETE("/profile/interests/:interestID", interest.remove)
	return router
}

// Health godoc
// @Summary      Проверка доступности сервиса
// @Description  Возвращает HTTP 200, если процесс принимает запросы.
// @Tags         health
// @Success      200  "Сервис доступен"
// @Router       /healthz [get]
func health(c *gin.Context) {
	c.Status(http.StatusOK)
}
