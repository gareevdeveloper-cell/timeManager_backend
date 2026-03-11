package domain

import (
	"time"

	"github.com/google/uuid"
)

// User — доменная модель пользователя (см. DOMAIN.md).
type User struct {
	ID              uuid.UUID
	Email           string
	PasswordHash    string
	OAuthProvider   string // google, yandex, github
	OAuthProviderID string // внешний ID пользователя у провайдера
	FirstName       string
	LastName        string
	MiddleName      string
	Birthday        *time.Time
	About           string // о себе
	Position        string // должность
	Role            string
	Status          string // активность аккаунта: active, inactive
	WorkStatus      string // рабочий статус: не работает, отдыхает, обед, отпуск, больничный, командировка
	AvatarURL       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UserStatus — допустимые статусы аккаунта пользователя.
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)

// WorkStatus — рабочий статус сотрудника.
const (
	WorkStatusWorking = "working" // работает
	WorkStatusResting     = "resting"       // отдыхает
	WorkStatusLunch       = "lunch"         // обед
	WorkStatusVacation    = "vacation"      // отпуск
	WorkStatusSickLeave   = "sick_leave"   // больничный
	WorkStatusBusinessTrip = "business_trip" // командировка
)

// ValidWorkStatuses — все допустимые значения WorkStatus.
var ValidWorkStatuses = map[string]bool{
	WorkStatusWorking: true,
	WorkStatusResting:     true,
	WorkStatusLunch:       true,
	WorkStatusVacation:    true,
	WorkStatusSickLeave:   true,
	WorkStatusBusinessTrip: true,
}

// UserRole — допустимые роли.
const (
	UserRoleUser  = "user"
	UserRoleAdmin = "admin"
)
