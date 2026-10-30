package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/gin-gonic/gin"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

func writeError(c *gin.Context, status int, code, message string) {
	var body ErrorResponse
	body.Error.Code = code
	body.Error.Message = message
	c.JSON(status, body)
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidUserID), errors.Is(err, domain.ErrInvalidProfile), errors.Is(err, domain.ErrInvalidInterestID):
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, domain.ErrProfileNotFound), errors.Is(err, domain.ErrInterestNotFound), errors.Is(err, domain.ErrUserNotFound):
		writeError(c, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, domain.ErrUsernameTaken):
		writeError(c, http.StatusConflict, "username_taken", err.Error())
	default:
		log.Printf("HTTP request failed: %v", err)
		writeError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func decodeJSON(c *gin.Context, dst any) bool {
	if c.ContentType() != "application/json" {
		writeError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
		return false
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeJSONError(c, err)
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err == nil {
			writeError(c, http.StatusBadRequest, "invalid_json", "request body must contain one JSON object")
		} else {
			writeJSONError(c, err)
		}
		return false
	}
	return true
}

func writeJSONError(c *gin.Context, err error) {
	var sizeErr *http.MaxBytesError
	if errors.As(err, &sizeErr) {
		writeError(c, http.StatusRequestEntityTooLarge, "body_too_large", "request body exceeds 1 MiB")
		return
	}
	writeError(c, http.StatusBadRequest, "invalid_json", "invalid JSON request body")
}
