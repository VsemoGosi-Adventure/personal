package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const userIDKey = "userID"

// UserIDMiddleware reads the identity supplied by the API gateway.
func UserIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		values := c.Request.Header.Values("X-User-ID")
		if len(values) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{
				"code": "missing_user_id", "message": "X-User-ID is required",
			}})
			return
		}
		id, err := uuid.Parse(values[0])
		if err != nil || id == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{
				"code": "invalid_user_id", "message": "X-User-ID must be a UUID",
			}})
			return
		}
		c.Set(userIDKey, id.String())
		c.Next()
	}
}

// UserID returns the UUID validated by UserIDMiddleware.
func UserID(c *gin.Context) string {
	return c.GetString(userIDKey)
}
