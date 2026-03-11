package team

import "errors"

var (
	ErrNotFound           = errors.New("team not found")
	ErrOrgNotFound        = errors.New("organization not found")
	ErrUserNotInOrg       = errors.New("user is not a member of the organization")
	ErrUserAlreadyInTeam  = errors.New("user is already in the team")
	ErrUserNotInTeam      = errors.New("user is not a member of the team")
	ErrInvalidAvatar      = errors.New("invalid avatar file")
)
