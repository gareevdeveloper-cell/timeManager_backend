package domain

// Роли для связей пользователь-организация, пользователь-команда, пользователь-проект.
const (
	RoleAdministrator = "administrator" // администратор — полный доступ
	RoleParticipant   = "participant"   // участник — участие в работе
	RoleUser          = "user"          // пользователь — базовый доступ
)

// ValidMemberRoles — допустимые роли участника.
var ValidMemberRoles = []string{RoleAdministrator, RoleParticipant, RoleUser}

// MemberWithRole — участник с ролью (для организаций, команд, проектов).
type MemberWithRole struct {
	User *User
	Role string
}
