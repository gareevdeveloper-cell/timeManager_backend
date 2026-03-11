package team

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/google/uuid"

	"testik/internal/domain"
	"testik/pkg/storage"
)

type mockStorage struct{}

func (mockStorage) Upload(ctx context.Context, path string, reader io.Reader, size int64, contentType string) (string, error) {
	return "https://example.com/" + path, nil
}
func (mockStorage) Delete(ctx context.Context, path string) error { return nil }
func (mockStorage) Get(ctx context.Context, path string) (io.ReadCloser, *storage.ObjectInfo, error) {
	return io.NopCloser(bytes.NewReader(nil)), &storage.ObjectInfo{ContentType: "image/jpeg", Size: 0}, nil
}

type mockTeamRepo struct {
	teams   map[uuid.UUID]*domain.Team
	members map[uuid.UUID]map[uuid.UUID]bool // teamID -> userIDs
}

func (m *mockTeamRepo) Create(ctx context.Context, t *domain.Team) error {
	if m.teams == nil {
		m.teams = make(map[uuid.UUID]*domain.Team)
	}
	m.teams[t.ID] = t
	return nil
}

func (m *mockTeamRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	return m.teams[id], nil
}

func (m *mockTeamRepo) GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]*domain.Team, error) {
	var result []*domain.Team
	for _, t := range m.teams {
		if t.OrganizationID == orgID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTeamRepo) Update(ctx context.Context, t *domain.Team) error { return nil }
func (m *mockTeamRepo) UpdateAvatarURL(ctx context.Context, teamID uuid.UUID, avatarURL string) error { return nil }
func (m *mockTeamRepo) Delete(ctx context.Context, id uuid.UUID) error    { return nil }

func (m *mockTeamRepo) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	if m.members == nil {
		m.members = make(map[uuid.UUID]map[uuid.UUID]bool)
	}
	if m.members[teamID] == nil {
		m.members[teamID] = make(map[uuid.UUID]bool)
	}
	m.members[teamID][userID] = true
	return nil
}

func (m *mockTeamRepo) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	if m.members != nil && m.members[teamID] != nil {
		delete(m.members[teamID], userID)
	}
	return nil
}

func (m *mockTeamRepo) IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	return m.members != nil && m.members[teamID] != nil && m.members[teamID][userID], nil
}

func (m *mockTeamRepo) GetMembers(ctx context.Context, teamID uuid.UUID) ([]*domain.MemberWithRole, error) {
	return nil, nil
}

type mockOrgRepo struct {
	orgs    map[uuid.UUID]*domain.Organization
	members map[uuid.UUID]uuid.UUID // userID -> orgID
}

func (m *mockOrgRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	return m.orgs[id], nil
}

func (m *mockOrgRepo) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	userOrg, ok := m.members[userID]
	return ok && userOrg == orgID, nil
}

func TestService_Create(t *testing.T) {
	orgID := uuid.New()
	creatorID := uuid.New()

	teamRepo := &mockTeamRepo{teams: make(map[uuid.UUID]*domain.Team)}
	orgRepo := &mockOrgRepo{
		orgs: map[uuid.UUID]*domain.Organization{
			orgID: {ID: orgID, Status: domain.OrganizationStatusActive},
		},
		members: map[uuid.UUID]uuid.UUID{creatorID: orgID},
	}

	svc := NewService(teamRepo, orgRepo, storage.Storage(mockStorage{}))
	ctx := context.Background()

	tm, err := svc.Create(ctx, "Backend Team", "Description", orgID, creatorID, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tm.Name != "Backend Team" {
		t.Errorf("expected name Backend Team, got %s", tm.Name)
	}
	if tm.OrganizationID != orgID {
		t.Errorf("organization_id mismatch")
	}
	if tm.CreatorID != creatorID {
		t.Errorf("creator_id mismatch")
	}
}

func TestService_Create_UserNotInOrg(t *testing.T) {
	orgID := uuid.New()
	creatorID := uuid.New()

	teamRepo := &mockTeamRepo{teams: make(map[uuid.UUID]*domain.Team)}
	orgRepo := &mockOrgRepo{
		orgs:    map[uuid.UUID]*domain.Organization{orgID: {ID: orgID}},
		members: map[uuid.UUID]uuid.UUID{}, // creator not in org
	}

	svc := NewService(teamRepo, orgRepo, storage.Storage(mockStorage{}))
	ctx := context.Background()

	_, err := svc.Create(ctx, "Team", "", orgID, creatorID, nil)
	if err != ErrUserNotInOrg {
		t.Errorf("expected ErrUserNotInOrg, got %v", err)
	}
}
