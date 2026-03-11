package domain

import (
	"time"

	"github.com/google/uuid"
)

// Team — доменная модель команды (см. DOMAIN.md).
// Команда принадлежит организации, в ней могут состоять несколько пользователей.
type Team struct {
	ID             uuid.UUID
	Name           string
	Description    string
	OrganizationID uuid.UUID
	CreatorID      uuid.UUID
	AvatarURL      string
	CreatedAt      time.Time
}

// TeamMember — членство пользователя в команде.
type TeamMember struct {
	ID        uuid.UUID
	TeamID    uuid.UUID
	UserID    uuid.UUID
	JoinedAt  time.Time
}
