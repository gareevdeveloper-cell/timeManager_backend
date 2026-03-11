package ws

import "encoding/json"

// UserWorkStatusChanged — событие смены статуса пользователя.
const ChannelUserStatus = "user:status"

// UserWorkStatusPayload — payload события смены статуса.
type UserWorkStatusPayload struct {
	Type    string `json:"type"`
	UserID  string `json:"user_id"`
	WorkStatus string `json:"work_status"`
}

// EncodeUserWorkStatusChanged кодирует событие смены статуса.
func EncodeUserWorkStatusChanged(userID, workStatus string) ([]byte, error) {
	return json.Marshal(UserWorkStatusPayload{
		Type:        "user.work_status_changed",
		UserID:      userID,
		WorkStatus:  workStatus,
	})
}
