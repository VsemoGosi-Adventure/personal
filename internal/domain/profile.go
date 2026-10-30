package domain

import "time"

type Profile struct {
	UserID    string
	Username  string
	Bio       *string
	AvatarURL *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ProfileUpdate replaces the editable fields of a user's profile.
type ProfileUpdate struct {
	Username  string
	Bio       *string
	AvatarURL *string
}
