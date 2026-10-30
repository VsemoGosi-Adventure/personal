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

type interestResponse struct {
	ID       int32  `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

func (h interestHandler) list(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, interestDTOs(items))
}

func (h interestHandler) listForUser(c *gin.Context) {
	items, err := h.service.ListForUser(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, interestDTOs(items))
}

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

func interestDTOs(items []domain.Interest) []interestResponse {
	result := make([]interestResponse, 0, len(items))
	for _, item := range items {
		result = append(result, interestResponse{ID: item.ID, Slug: item.Slug, Name: item.Name, Category: item.Category})
	}
	return result
}
