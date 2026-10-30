package service

import (
	"errors"

	"github.com/VsemoGosi-Adventure/personal/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func parseUserID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return pgtype.UUID{}, domain.ErrInvalidUserID
	}
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}, nil
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch {
	case pgErr.Code == "23505" && pgErr.ConstraintName == "profiles_username_key":
		return domain.ErrUsernameTaken
	case pgErr.Code == "23503" && pgErr.ConstraintName == "user_interests_interest_id_fkey":
		return domain.ErrInterestNotFound
	case pgErr.Code == "23503":
		return domain.ErrUserNotFound
	default:
		return err
	}
}

func textValue(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
