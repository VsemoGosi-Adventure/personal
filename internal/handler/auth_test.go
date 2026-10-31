package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/VsemoGosi-Adventure/personal/internal/service"
	"github.com/gin-gonic/gin"
)

type stubAuth struct{ service.AuthService }

func (stubAuth) Register(_ context.Context, _ domain.RegisterInput) (domain.AuthTokens, error) {
	return domain.AuthTokens{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 900}, nil
}

func (stubAuth) Verify(token string) (string, error) {
	if token != "valid" {
		return "", domain.ErrInvalidToken
	}
	return "f65a8c84-3cc9-48d5-a5da-71a022f553ae", nil
}

func TestAuthRoutesDoNotRequireUserHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(nil, nil, stubAuth{})
	register := httptest.NewRequest(http.MethodPost, "/v1/auth/register", strings.NewReader(`{"email":"a@example.com","password":"password123","username":"alice"}`))
	register.Header.Set("Content-Type", "application/json")
	result := httptest.NewRecorder()
	router.ServeHTTP(result, register)
	if result.Code != http.StatusCreated || !strings.Contains(result.Body.String(), `"expires_in":900`) {
		t.Fatalf("register: status=%d body=%s", result.Code, result.Body.String())
	}

	for _, tc := range []struct {
		authorization string
		status        int
	}{
		{"Bearer valid", http.StatusOK},
		{"Bearer invalid", http.StatusUnauthorized},
		{"", http.StatusUnauthorized},
	} {
		request := httptest.NewRequest(http.MethodGet, "/v1/auth/verify", nil)
		request.Header.Set("Authorization", tc.authorization)
		result := httptest.NewRecorder()
		router.ServeHTTP(result, request)
		if result.Code != tc.status {
			t.Errorf("verify %q: got %d, want %d", tc.authorization, result.Code, tc.status)
		}
		if tc.status == http.StatusOK && result.Header().Get("X-User-ID") == "" {
			t.Error("valid token did not set X-User-ID")
		}
	}
}
