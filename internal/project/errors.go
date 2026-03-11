package project

import "errors"

var (
	ErrProjectNotFound  = errors.New("project not found")
	ErrTeamNotFound    = errors.New("team not found")
	ErrUserNotInTeam   = errors.New("user is not a member of the team")
	ErrKeyAlreadyExists = errors.New("project key already exists")
	ErrInvalidKey      = errors.New("invalid project key format")
	ErrForbidden       = errors.New("access denied to project")
	ErrTaskNotFound    = errors.New("task not found")
	ErrInvalidStatus   = errors.New("invalid task status")
	ErrInvalidPriority = errors.New("invalid task priority")
	ErrInvalidRequest   = errors.New("invalid request data")
	ErrStatusNotFound   = errors.New("status not found")
	ErrStatusKeyExists  = errors.New("status key already exists in project")
)
