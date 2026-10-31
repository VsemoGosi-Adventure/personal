package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/VsemoGosi-Adventure/personal/internal/db"
	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestAccessTokenVerification(t *testing.T) {
	id := uuid.New()
	svc := &authService{secret: []byte("a-32-byte-secret-for-auth-unit-tests")}
	token, err := svc.newAccessToken(pgtype.UUID{Bytes: [16]byte(id), Valid: true}, "user")
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Verify(token)
	if err != nil || got != id.String() {
		t.Fatalf("verify valid token: id=%q err=%v", got, err)
	}
	other := &authService{secret: []byte("another-32-byte-secret-for-testing")}
	if _, err := other.Verify(token); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("token with wrong signature: %v", err)
	}
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims{
		UserID: id.String(), Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{Issuer: "personal", ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute))},
	}).SignedString(svc.secret)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Verify(expired); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expired token: %v", err)
	}
}

func TestRefreshTokenRotatesOnce(t *testing.T) {
	oldToken, oldHash, err := newRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	store := &refreshStore{
		userID:  pgtype.UUID{Bytes: [16]byte(uuid.New()), Valid: true},
		hash:    oldHash,
		expires: time.Now().Add(time.Hour),
	}
	svc := &authService{q: store, secret: []byte("a-32-byte-secret-for-auth-unit-tests")}
	result, err := svc.Refresh(context.Background(), oldToken)
	if err != nil {
		t.Fatal(err)
	}
	if result.RefreshToken == oldToken || result.ExpiresIn != 900 {
		t.Fatalf("unexpected token pair: %+v", result)
	}
	if _, err := svc.Refresh(context.Background(), oldToken); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("reused refresh token: %v", err)
	}
	if _, err := svc.Refresh(context.Background(), result.RefreshToken); err != nil {
		t.Fatalf("new refresh token: %v", err)
	}
}

func TestConcurrentRefreshHasOneWinner(t *testing.T) {
	token, hash, err := newRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	store := &refreshStore{
		userID:  pgtype.UUID{Bytes: [16]byte(uuid.New()), Valid: true},
		hash:    hash,
		expires: time.Now().Add(time.Hour),
	}
	svc := &authService{q: store, secret: []byte("a-32-byte-secret-for-auth-unit-tests")}
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			_, err := svc.Refresh(context.Background(), token)
			results <- err
		}()
	}
	var succeeded, rejected int
	for i := 0; i < 2; i++ {
		switch err := <-results; {
		case err == nil:
			succeeded++
		case errors.Is(err, domain.ErrInvalidToken):
			rejected++
		default:
			t.Fatalf("unexpected refresh error: %v", err)
		}
	}
	if succeeded != 1 || rejected != 1 {
		t.Fatalf("success=%d rejected=%d", succeeded, rejected)
	}
}

type refreshStore struct {
	db.Querier
	mu      sync.Mutex
	userID  pgtype.UUID
	hash    string
	expires time.Time
}

func (s *refreshStore) GetCredentialsByRefreshHash(_ context.Context, hash pgtype.Text) (db.GetCredentialsByRefreshHashRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if hash.String != s.hash {
		return db.GetCredentialsByRefreshHashRow{}, pgx.ErrNoRows
	}
	return db.GetCredentialsByRefreshHashRow{
		UserID: s.userID, Role: "user", RefreshExpiresAt: pgtype.Timestamptz{Time: s.expires, Valid: true},
	}, nil
}

func (s *refreshStore) RotateRefreshToken(_ context.Context, p db.RotateRefreshTokenParams) (db.RotateRefreshTokenRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.RefreshTokenHash.String != s.hash || !time.Now().Before(s.expires) {
		return db.RotateRefreshTokenRow{}, pgx.ErrNoRows
	}
	s.hash = p.RefreshTokenHash_2.String
	s.expires = p.RefreshExpiresAt.Time
	return db.RotateRefreshTokenRow{UserID: s.userID, Role: "user"}, nil
}
