package service

import (
	"context"
	"errors"

	"github.com/VsemoGosi-Adventure/personal/internal/db"
	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/jackc/pgx/v5"
)

type interestService struct {
	queries db.Querier
}

func NewInterestService(queries db.Querier) InterestService {
	return &interestService{queries: queries}
}

func (s *interestService) List(ctx context.Context) ([]domain.Interest, error) {
	rows, err := s.queries.ListInterests(ctx)
	if err != nil {
		return nil, err
	}
	interests := make([]domain.Interest, 0, len(rows))
	for _, row := range rows {
		interests = append(interests, domain.Interest{ID: row.ID, Slug: row.Slug, Name: row.Name, Category: row.Category})
	}
	return interests, nil
}

func (s *interestService) ListForUser(ctx context.Context, userID string) ([]domain.Interest, error) {
	id, err := parseUserID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListUserInterests(ctx, id)
	if err != nil {
		return nil, err
	}
	interests := make([]domain.Interest, 0, len(rows))
	for _, row := range rows {
		interests = append(interests, domain.Interest{ID: row.ID, Slug: row.Slug, Name: row.Name, Category: row.Category})
	}
	return interests, nil
}

func (s *interestService) Add(ctx context.Context, userID string, interestID int32) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}
	if interestID <= 0 {
		return domain.ErrInvalidInterestID
	}
	if _, err := s.queries.GetInterest(ctx, interestID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrInterestNotFound
		}
		return err
	}
	return mapWriteError(s.queries.AddUserInterest(ctx, db.AddUserInterestParams{UserID: id, InterestID: interestID}))
}

func (s *interestService) Remove(ctx context.Context, userID string, interestID int32) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}
	if interestID <= 0 {
		return domain.ErrInvalidInterestID
	}
	return s.queries.RemoveUserInterest(ctx, db.RemoveUserInterestParams{UserID: id, InterestID: interestID})
}
