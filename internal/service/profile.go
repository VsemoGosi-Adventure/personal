package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/VsemoGosi-Adventure/personal/internal/db"
	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/jackc/pgx/v5"
)

type profileService struct {
	queries db.Querier
}

func NewProfileService(queries db.Querier) ProfileService {
	return &profileService{queries: queries}
}

func (s *profileService) Get(ctx context.Context, userID string) (domain.Profile, error) {
	id, err := parseUserID(userID)
	if err != nil {
		return domain.Profile{}, err
	}
	row, err := s.queries.GetProfile(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Profile{}, domain.ErrProfileNotFound
	}
	if err != nil {
		return domain.Profile{}, err
	}
	return domain.Profile{
		UserID:    userID,
		Username:  row.Username,
		Bio:       optionalText(row.Bio),
		AvatarURL: optionalText(row.AvatarUrl),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (s *profileService) Upsert(ctx context.Context, userID string, update domain.ProfileUpdate) (domain.Profile, error) {
	id, err := parseUserID(userID)
	if err != nil {
		return domain.Profile{}, err
	}
	username := strings.TrimSpace(update.Username)
	if username == "" || utf8.RuneCountInString(username) > 50 {
		return domain.Profile{}, domain.ErrInvalidProfile
	}
	row, err := s.queries.UpsertProfile(ctx, db.UpsertProfileParams{
		UserID: id, Username: username, Bio: textValue(update.Bio), AvatarUrl: textValue(update.AvatarURL),
	})
	if err != nil {
		return domain.Profile{}, mapWriteError(err)
	}
	return domain.Profile{
		UserID:    userID,
		Username:  row.Username,
		Bio:       optionalText(row.Bio),
		AvatarURL: optionalText(row.AvatarUrl),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}
