package organization

import "errors"

var (
	ErrSlugConflict      = errors.New("slug already exists")
	ErrNotFound          = errors.New("organization not found")
	ErrArchived          = errors.New("organization is archived")
	ErrUserAlreadyInOrg  = errors.New("user is already in an organization")
	ErrUserNotInOrg      = errors.New("user is not a member of this organization")
	ErrInvalidAvatar     = errors.New("invalid avatar file")
)
