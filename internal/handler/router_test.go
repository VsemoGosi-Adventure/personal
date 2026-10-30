package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/gin-gonic/gin"
)

const testUserID = "caa6a7d0-cf47-424c-8762-b6d00e7c23a4"

type profileStub struct {
	get    func(context.Context, string) (domain.Profile, error)
	upsert func(context.Context, string, domain.ProfileUpdate) (domain.Profile, error)
}

func (s profileStub) Get(ctx context.Context, userID string) (domain.Profile, error) {
	return s.get(ctx, userID)
}

func (s profileStub) Upsert(ctx context.Context, userID string, update domain.ProfileUpdate) (domain.Profile, error) {
	return s.upsert(ctx, userID, update)
}

type interestStub struct {
	list        func(context.Context) ([]domain.Interest, error)
	listForUser func(context.Context, string) ([]domain.Interest, error)
	add         func(context.Context, string, int32) error
	remove      func(context.Context, string, int32) error
}

func (s interestStub) List(ctx context.Context) ([]domain.Interest, error) {
	return s.list(ctx)
}

func (s interestStub) ListForUser(ctx context.Context, userID string) ([]domain.Interest, error) {
	return s.listForUser(ctx, userID)
}

func (s interestStub) Add(ctx context.Context, userID string, interestID int32) error {
	return s.add(ctx, userID, interestID)
}

func (s interestStub) Remove(ctx context.Context, userID string, interestID int32) error {
	return s.remove(ctx, userID, interestID)
}

func perform(router http.Handler, method, path, body string, authorized bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if authorized {
		req.Header.Set("X-User-ID", testUserID)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func TestUserIDMiddlewareAndHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(profileStub{get: func(_ context.Context, id string) (domain.Profile, error) {
		if id != testUserID {
			t.Fatalf("unexpected user ID %q", id)
		}
		return domain.Profile{UserID: id, Username: "Alice"}, nil
	}}, interestStub{})
	if got := perform(router, http.MethodGet, "/healthz", "", false).Code; got != http.StatusOK {
		t.Fatalf("health status = %d", got)
	}
	if got := perform(router, http.MethodGet, "/v1/profile", "", false).Code; got != http.StatusUnauthorized {
		t.Fatalf("missing ID status = %d", got)
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/profile", nil)
	req.Header.Set("X-User-ID", "bad")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("invalid ID status = %d", response.Code)
	}
	response = perform(router, http.MethodGet, "/v1/profile", "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("profile status = %d: %s", response.Code, response.Body.String())
	}
	var profile map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}
	if profile["userId"] != testUserID || profile["username"] != "Alice" {
		t.Fatalf("unexpected profile JSON: %v", profile)
	}
}

func TestProfilePutBindingAndConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(profileStub{upsert: func(_ context.Context, id string, update domain.ProfileUpdate) (domain.Profile, error) {
		if id != testUserID || update.Username != "Alice" || update.Bio == nil || *update.Bio != "Hello" {
			t.Fatalf("unexpected update: id=%q update=%+v", id, update)
		}
		return domain.Profile{}, domain.ErrUsernameTaken
	}}, interestStub{})
	if got := perform(router, http.MethodPut, "/v1/profile", `{broken`, true).Code; got != http.StatusBadRequest {
		t.Fatalf("invalid JSON status = %d", got)
	}
	if got := perform(router, http.MethodPut, "/v1/profile", `{"username":"Alice","extra":1}`, true).Code; got != http.StatusBadRequest {
		t.Fatalf("unknown field status = %d", got)
	}
	response := perform(router, http.MethodPut, "/v1/profile", `{"username":"Alice","bio":"Hello"}`, true)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"username_taken"`) {
		t.Fatalf("conflict response = %d %s", response.Code, response.Body.String())
	}
}

func TestInterestRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	interest := domain.Interest{ID: 7, Slug: "hiking", Name: "Hiking", Category: "outdoor"}
	added, removed := false, false
	router := NewRouter(profileStub{}, interestStub{
		list: func(context.Context) ([]domain.Interest, error) { return []domain.Interest{interest}, nil },
		listForUser: func(_ context.Context, id string) ([]domain.Interest, error) {
			if id != testUserID {
				t.Fatalf("unexpected user ID %q", id)
			}
			return []domain.Interest{interest}, nil
		},
		add: func(_ context.Context, id string, interestID int32) error {
			added = id == testUserID && interestID == 7
			return nil
		},
		remove: func(_ context.Context, id string, interestID int32) error {
			removed = id == testUserID && interestID == 7
			return nil
		},
	})
	for _, path := range []string{"/v1/interests", "/v1/profile/interests"} {
		response := perform(router, http.MethodGet, path, "", true)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"slug":"hiking"`) {
			t.Fatalf("list %s = %d %s", path, response.Code, response.Body.String())
		}
	}
	path := "/v1/profile/interests/7"
	if got := perform(router, http.MethodPut, path, "", true).Code; got != http.StatusNoContent || !added {
		t.Fatalf("add status = %d, called = %v", got, added)
	}
	if got := perform(router, http.MethodDelete, path, "", true).Code; got != http.StatusNoContent || !removed {
		t.Fatalf("remove status = %d, called = %v", got, removed)
	}
	if got := perform(router, http.MethodPut, "/v1/profile/interests/nope", "", true).Code; got != http.StatusBadRequest {
		t.Fatalf("invalid interest ID status = %d", got)
	}
}

func TestServiceErrorsDoNotLeakInternals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(profileStub{get: func(context.Context, string) (domain.Profile, error) {
		return domain.Profile{}, errors.New("database password leaked")
	}}, interestStub{})
	response := perform(router, http.MethodGet, "/v1/profile", "", true)
	if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), "password") {
		t.Fatalf("internal error response = %d %s", response.Code, response.Body.String())
	}
}
