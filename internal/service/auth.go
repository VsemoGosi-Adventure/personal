package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/VsemoGosi-Adventure/personal/internal/db"
	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const accessLifetime = 15 * time.Minute
const refreshLifetime = 30 * 24 * time.Hour
const passwordCost = 10

type authService struct {
	pool   *pgxpool.Pool
	q      db.Querier
	secret []byte
}

type accessClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(pool *pgxpool.Pool, q db.Querier, secret string) AuthService {
	return &authService{pool: pool, q: q, secret: []byte(secret)}
}

func (s *authService) Register(ctx context.Context, input domain.RegisterInput) (domain.AuthTokens, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil || len(input.Password) < 8 || len(input.Password) > 72 {
		return domain.AuthTokens{}, domain.ErrInvalidAuthInput
	}
	username := strings.TrimSpace(input.Username)
	if username == "" || len(username) > 50 {
		return domain.AuthTokens{}, domain.ErrInvalidAuthInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), passwordCost)
	if err != nil {
		return domain.AuthTokens{}, fmt.Errorf("hash password: %w", err)
	}
	refresh, refreshHash, err := newRefreshToken()
	if err != nil {
		return domain.AuthTokens{}, err
	}
	expires := time.Now().Add(refreshLifetime)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.AuthTokens{}, fmt.Errorf("begin registration: %w", err)
	}
	defer tx.Rollback(ctx)
	q := db.New(tx)
	credential, err := q.CreateCredentials(ctx, db.CreateCredentialsParams{
		Email: email, PasswordHash: string(hash),
		RefreshTokenHash: pgtype.Text{String: refreshHash, Valid: true},
		RefreshExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
	})
	if err != nil {
		return domain.AuthTokens{}, mapAuthWriteError(err)
	}
	if err := q.CreateInitialProfile(ctx, db.CreateInitialProfileParams{UserID: credential.UserID, Username: username}); err != nil {
		return domain.AuthTokens{}, mapAuthWriteError(err)
	}
	access, err := s.newAccessToken(credential.UserID, credential.Role)
	if err != nil {
		return domain.AuthTokens{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.AuthTokens{}, mapAuthWriteError(err)
	}
	return domain.AuthTokens{AccessToken: access, RefreshToken: refresh, ExpiresIn: int(accessLifetime.Seconds())}, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (domain.AuthTokens, error) {
	normalized, err := normalizeEmail(email)
	if err != nil || password == "" {
		return domain.AuthTokens{}, domain.ErrInvalidCredentials
	}
	credential, err := s.q.GetCredentialsByEmail(ctx, normalized)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AuthTokens{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return domain.AuthTokens{}, fmt.Errorf("get credentials: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(credential.PasswordHash), []byte(password)) != nil {
		return domain.AuthTokens{}, domain.ErrInvalidCredentials
	}
	if credential.IsBlocked {
		return domain.AuthTokens{}, domain.ErrAccountBlocked
	}
	refresh, refreshHash, err := newRefreshToken()
	if err != nil {
		return domain.AuthTokens{}, err
	}
	access, err := s.newAccessToken(credential.UserID, credential.Role)
	if err != nil {
		return domain.AuthTokens{}, err
	}
	count, err := s.q.UpdateRefreshToken(ctx, db.UpdateRefreshTokenParams{
		UserID:           credential.UserID,
		RefreshTokenHash: pgtype.Text{String: refreshHash, Valid: true},
		RefreshExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(refreshLifetime), Valid: true},
	})
	if err != nil {
		return domain.AuthTokens{}, fmt.Errorf("save refresh token: %w", err)
	}
	if count == 0 {
		return domain.AuthTokens{}, domain.ErrAccountBlocked
	}
	return domain.AuthTokens{AccessToken: access, RefreshToken: refresh, ExpiresIn: int(accessLifetime.Seconds())}, nil
}

func (s *authService) Refresh(ctx context.Context, oldToken string) (domain.AuthTokens, error) {
	oldHash, err := hashRefreshToken(oldToken)
	if err != nil {
		return domain.AuthTokens{}, err
	}
	credential, err := s.q.GetCredentialsByRefreshHash(ctx, pgtype.Text{String: oldHash, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AuthTokens{}, domain.ErrInvalidToken
	}
	if err != nil {
		return domain.AuthTokens{}, fmt.Errorf("get refresh token: %w", err)
	}
	if credential.IsBlocked || !credential.RefreshExpiresAt.Valid || !time.Now().Before(credential.RefreshExpiresAt.Time) {
		return domain.AuthTokens{}, domain.ErrInvalidToken
	}
	refresh, newHash, err := newRefreshToken()
	if err != nil {
		return domain.AuthTokens{}, err
	}
	rotated, err := s.q.RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
		RefreshTokenHash:   pgtype.Text{String: oldHash, Valid: true},
		RefreshTokenHash_2: pgtype.Text{String: newHash, Valid: true},
		RefreshExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(refreshLifetime), Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AuthTokens{}, domain.ErrInvalidToken
	}
	if err != nil {
		return domain.AuthTokens{}, fmt.Errorf("rotate refresh token: %w", err)
	}
	access, err := s.newAccessToken(rotated.UserID, rotated.Role)
	if err != nil {
		return domain.AuthTokens{}, err
	}
	return domain.AuthTokens{AccessToken: access, RefreshToken: refresh, ExpiresIn: int(accessLifetime.Seconds())}, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	hash, err := hashRefreshToken(refreshToken)
	if err != nil {
		return err
	}
	_, err = s.q.InvalidateRefreshToken(ctx, pgtype.Text{String: hash, Valid: true})
	if err != nil {
		return fmt.Errorf("invalidate refresh token: %w", err)
	}
	return nil
}

func (s *authService) Verify(accessToken string) (string, error) {
	claims := &accessClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(*jwt.Token) (any, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithIssuer("personal"))
	if err != nil || !token.Valid || claims.UserID == "" || claims.Role == "" {
		return "", domain.ErrInvalidToken
	}
	parsedID, err := uuid.Parse(claims.UserID)
	if err != nil || parsedID == uuid.Nil {
		return "", domain.ErrInvalidToken
	}
	return claims.UserID, nil
}

func (s *authService) newAccessToken(userID pgtype.UUID, role string) (string, error) {
	if !userID.Valid || role == "" {
		return "", fmt.Errorf("invalid credentials for access token")
	}
	now := time.Now()
	claims := accessClaims{
		UserID: uuid.UUID(userID.Bytes).String(), Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "personal", IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(accessLifetime)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func newRefreshToken() (string, string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	token := hex.EncodeToString(raw[:])
	hash := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(hash[:]), nil
}

func hashRefreshToken(token string) (string, error) {
	if len(token) != 64 {
		return "", domain.ErrInvalidToken
	}
	if _, err := hex.DecodeString(token); err != nil {
		return "", domain.ErrInvalidToken
	}
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:]), nil
}

func normalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	if len(email) == 0 || len(email) > 255 {
		return "", domain.ErrInvalidAuthInput
	}
	parsed, err := mail.ParseAddress(email)
	_, domainPart, hasAt := strings.Cut(email, "@")
	if err != nil || parsed.Address != email || !hasAt || !strings.Contains(domainPart, ".") {
		return "", domain.ErrInvalidAuthInput
	}
	return email, nil
}

func mapAuthWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "credentials_email_key":
			return domain.ErrEmailTaken
		case "profiles_username_key":
			return domain.ErrUsernameTaken
		}
	}
	return fmt.Errorf("write credentials: %w", err)
}
