package organization

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

type mockRepo struct {
	orgs       map[string]*domain.Organization
	orgsByID   map[uuid.UUID]*domain.Organization
	createErr  error
	existsBy   map[string]bool
	members    map[uuid.UUID]uuid.UUID // userID -> orgID
}

func (m *mockRepo) Create(ctx context.Context, o *domain.Organization) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.orgs[o.Slug] = o
	if m.orgsByID != nil {
		m.orgsByID[o.ID] = o
	}
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	if m.orgsByID != nil {
		return m.orgsByID[id], nil
	}
	for _, o := range m.orgs {
		if o.ID == id {
			return o, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) Update(ctx context.Context, o *domain.Organization) error { return nil }
func (m *mockRepo) UpdateStatus(ctx context.Context, o *domain.Organization) error { return nil }
func (m *mockRepo) UpdateAvatarURL(ctx context.Context, orgID uuid.UUID, avatarURL string) error { return nil }

func (m *mockRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return m.existsBy[slug], nil
}

func (m *mockRepo) AddMember(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	if m.members == nil {
		m.members = make(map[uuid.UUID]uuid.UUID)
	}
	m.members[userID] = orgID
	return nil
}

func (m *mockRepo) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error { return nil }

func (m *mockRepo) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	if m.members == nil {
		return false, nil
	}
	userOrg, ok := m.members[userID]
	return ok && userOrg == orgID, nil
}

func (m *mockRepo) GetMemberOrganization(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	if m.members == nil {
		return nil, nil
	}
	orgID, ok := m.members[userID]
	if !ok {
		return nil, nil
	}
	return &orgID, nil
}

func (m *mockRepo) ListByMember(ctx context.Context, userID uuid.UUID) ([]*domain.Organization, error) {
	var orgs []*domain.Organization
	for _, o := range m.orgs {
		if m.members != nil && m.members[userID] == o.ID {
			orgs = append(orgs, o)
		}
	}
	return orgs, nil
}

func (m *mockRepo) ListMembers(ctx context.Context, orgID uuid.UUID) ([]*domain.MemberWithRole, error) {
	return nil, nil
}

func TestService_Create(t *testing.T) {
	repo := &mockRepo{
		orgs:     make(map[string]*domain.Organization),
		orgsByID: make(map[uuid.UUID]*domain.Organization),
		existsBy: make(map[string]bool),
		members:  make(map[uuid.UUID]uuid.UUID),
	}
	svc := NewService(repo, storage.Storage(mockStorage{}))
	ctx := context.Background()
	ownerID := uuid.New()

	o, err := svc.Create(ctx, "My Company", ownerID, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if o.Name != "My Company" {
		t.Errorf("expected name My Company, got %s", o.Name)
	}
	if o.Slug != "my-company" {
		t.Errorf("expected slug my-company, got %s", o.Slug)
	}
	if o.OwnerID != ownerID {
		t.Errorf("owner_id mismatch")
	}
}

func TestService_Create_SlugConflict(t *testing.T) {
	repo := &mockRepo{
		orgs:     make(map[string]*domain.Organization),
		orgsByID: make(map[uuid.UUID]*domain.Organization),
		existsBy: map[string]bool{"my-company": true},
		members:  make(map[uuid.UUID]uuid.UUID),
	}
	svc := NewService(repo, storage.Storage(mockStorage{}))
	ctx := context.Background()

	o, err := svc.Create(ctx, "My Company", uuid.New(), nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if o.Slug != "my-company-2" {
		t.Errorf("expected slug my-company-2 for conflict, got %s", o.Slug)
	}
}
