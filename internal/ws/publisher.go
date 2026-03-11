package ws

// WorkStatusPublisher — публикация события смены статуса пользователя.
type WorkStatusPublisher interface {
	PublishWorkStatusChanged(userID, workStatus string)
}

// HubWorkStatusPublisher реализует WorkStatusPublisher через Hub.
type HubWorkStatusPublisher struct {
	hub *Hub
}

// NewHubWorkStatusPublisher создаёт publisher.
func NewHubWorkStatusPublisher(hub *Hub) *HubWorkStatusPublisher {
	return &HubWorkStatusPublisher{hub: hub}
}

// PublishWorkStatusChanged рассылает событие смены статуса подписчикам канала user:status.
func (p *HubWorkStatusPublisher) PublishWorkStatusChanged(userID, workStatus string) {
	payload, err := EncodeUserWorkStatusChanged(userID, workStatus)
	if err != nil {
		return
	}
	p.hub.Broadcast(ChannelUserStatus, payload)
}
