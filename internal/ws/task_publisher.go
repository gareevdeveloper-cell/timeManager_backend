package ws

import "github.com/google/uuid"

// HubTaskPublisher публикует события задач в каналы проекта.
type HubTaskPublisher struct {
	hub *Hub
}

// NewHubTaskPublisher создаёт publisher задач.
func NewHubTaskPublisher(hub *Hub) *HubTaskPublisher {
	return &HubTaskPublisher{hub: hub}
}

// PublishTaskChanged рассылает событие в канал {projectId}:taskChange.
func (p *HubTaskPublisher) PublishTaskChanged(projectID, taskID uuid.UUID, action string) {
	payload, err := EncodeProjectTaskChange(projectID, taskID, action)
	if err != nil {
		return
	}
	p.hub.Broadcast(ChannelProjectTaskChange(projectID), payload)
}

// PublishTaskStatusChanged рассылает событие в канал {projectId}:taskChangeStatus.
func (p *HubTaskPublisher) PublishTaskStatusChanged(projectID, taskID uuid.UUID, statusKey string, statusID uuid.UUID) {
	payload, err := EncodeProjectTaskStatusChange(projectID, taskID, statusKey, statusID)
	if err != nil {
		return
	}
	p.hub.Broadcast(ChannelProjectTaskChangeStatus(projectID), payload)
}
