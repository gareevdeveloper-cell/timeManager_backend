package domain

import "github.com/google/uuid"

// Роли для связей пользователь-организация, пользователь-команда, пользователь-проект.
const (
	RoleAdministrator = "administrator" // администратор — полный доступ
	RoleParticipant   = "participant"   // участник — участие в работе
	RoleUser          = "user"          // пользователь — базовый доступ
)

// ValidMemberRoles — допустимые роли участника.
var ValidMemberRoles = []string{RoleAdministrator, RoleParticipant, RoleUser}

// MemberCurrentTask — краткие данные текущей задачи участника (для списков members без отдельного GET /tasks/:id).
type MemberCurrentTask struct {
	ID        uuid.UUID
	Title     string
	ProjectID uuid.UUID
}

// MemberWithRole — участник с ролью (для организаций, команд, проектов).
type MemberWithRole struct {
	User        *User
	Role        string
	CurrentTask *MemberCurrentTask // nil, если нет текущей задачи
}
