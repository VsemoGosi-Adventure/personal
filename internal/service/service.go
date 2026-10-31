package service

import (
	"context"

	"github.com/VsemoGosi-Adventure/personal/internal/domain"
)

type ProfileService interface {
	Get(ctx context.Context, userID string) (domain.Profile, error)
	Upsert(ctx context.Context, userID string, update domain.ProfileUpdate) (domain.Profile, error)
}

type InterestService interface {
	List(ctx context.Context) ([]domain.Interest, error)
	ListForUser(ctx context.Context, userID string) ([]domain.Interest, error)
	Add(ctx context.Context, userID string, interestID int32) error
	Remove(ctx context.Context, userID string, interestID int32) error
}

type AuthService interface {
	Register(ctx context.Context, input domain.RegisterInput) (domain.AuthTokens, error)
	Login(ctx context.Context, email, password string) (domain.AuthTokens, error)
	Refresh(ctx context.Context, refreshToken string) (domain.AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	Verify(accessToken string) (string, error)
}

// SettingsService is a contract only until settings storage is defined in the schema.
type SettingsService interface {
	Get(ctx context.Context, userID string) (domain.PrivacySettings, error)
	Update(ctx context.Context, userID string, update domain.PrivacySettingsUpdate) (domain.PrivacySettings, error)
}
