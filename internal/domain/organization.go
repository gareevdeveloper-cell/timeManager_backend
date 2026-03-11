package domain

import (
	"time"

	"github.com/google/uuid"
)

// Organization — доменная модель организации (см. DOMAIN.md).
type Organization struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	OwnerID   uuid.UUID
	Status    string
	AvatarURL string // URL аватарки в MinIO
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OrganizationMember — членство пользователя в организации.
type OrganizationMember struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	UserID        uuid.UUID
	JoinedAt      time.Time
}

// OrganizationStatus — допустимые статусы организации.
const (
	OrganizationStatusActive   = "active"
	OrganizationStatusArchived = "archived"
)
