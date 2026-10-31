package handler

import (
	"net/http"
	"strconv"

	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/VsemoGosi-Adventure/personal/internal/middleware"
	"github.com/VsemoGosi-Adventure/personal/internal/service"
	"github.com/gin-gonic/gin"
)

type interestHandler struct {
	service service.InterestService
}

type InterestResponse struct {
	ID       int32  `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// ListInterests godoc
// @Summary      Список доступных интересов
// @Description  Возвращает справочник интересов.
// @Tags         interests
// @Produce      json
// @Security     XUserID
// @Success      200  {array}   InterestResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /v1/interests [get]
func (h interestHandler) list(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, interestDTOs(items))
}

// ListUserInterests godoc
// @Summary      Список интересов пользователя
// @Description  Возвращает интересы пользователя из заголовка X-User-ID.
// @Tags         interests
// @Produce      json
// @Security     XUserID
// @Success      200  {array}   InterestResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /v1/profile/interests [get]
func (h interestHandler) listForUser(c *gin.Context) {
	items, err := h.service.ListForUser(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, interestDTOs(items))
}

// AddUserInterest godoc
// @Summary      Добавление интереса
// @Description  Добавляет интерес пользователю из заголовка X-User-ID. Повторный запрос безопасен.
// @Tags         interests
// @Produce      json
// @Security     XUserID
// @Param        interestID  path  int  true  "ID интереса"
// @Success      204  "Интерес добавлен"
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /v1/profile/interests/{interestID} [put]
func (h interestHandler) add(c *gin.Context) {
	interestID, ok := parseInterestID(c)
	if !ok {
		return
	}
	if err := h.service.Add(c.Request.Context(), middleware.UserID(c), interestID); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveUserInterest godoc
// @Summary      Удаление интереса
// @Description  Удаляет интерес пользователя из заголовка X-User-ID. Повторный запрос безопасен.
// @Tags         interests
// @Produce      json
// @Security     XUserID
// @Param        interestID  path  int  true  "ID интереса"
// @Success      204  "Интерес удалён"
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /v1/profile/interests/{interestID} [delete]
func (h interestHandler) remove(c *gin.Context) {
	interestID, ok := parseInterestID(c)
	if !ok {
		return
	}
	if err := h.service.Remove(c.Request.Context(), middleware.UserID(c), interestID); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseInterestID(c *gin.Context) (int32, bool) {
	value, err := strconv.ParseInt(c.Param("interestID"), 10, 32)
	if err != nil || value <= 0 {
		writeError(c, http.StatusBadRequest, "invalid_interest_id", "interestID must be a positive integer")
		return 0, false
	}
	return int32(value), true
}

func interestDTOs(items []domain.Interest) []InterestResponse {
	result := make([]InterestResponse, 0, len(items))
	for _, item := range items {
		result = append(result, InterestResponse{ID: item.ID, Slug: item.Slug, Name: item.Name, Category: item.Category})
	}
	return result
}
