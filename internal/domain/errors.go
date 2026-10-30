package domain

import "errors"

var (
	ErrInvalidUserID     = errors.New("invalid user ID")
	ErrInvalidProfile    = errors.New("invalid profile")
	ErrInvalidInterestID = errors.New("invalid interest ID")
	ErrProfileNotFound   = errors.New("profile not found")
	ErrInterestNotFound  = errors.New("interest not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrUsernameTaken     = errors.New("username already taken")
)
