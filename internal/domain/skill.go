package domain

import "github.com/google/uuid"

// Skill — скилл (навык) пользователя.
type Skill struct {
	ID   uuid.UUID
	Name string
}
