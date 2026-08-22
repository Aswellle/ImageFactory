package service

import (
	"context"
	"fmt"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/project"
	"github.com/imageforge/imageforge/internal/pkg/errors"
)

// ProjectService handles project business logic.
type ProjectService struct {
	db *ent.Client
}

// NewProjectService builds a ProjectService.
func NewProjectService(db *ent.Client) *ProjectService {
	return &ProjectService{db: db}
}

// ProjectCreateInput is the input for creating a project.
type ProjectCreateInput struct {
	UserID      int64
	Name        string
	Description string
}

// Create creates a new project.
func (s *ProjectService) Create(ctx context.Context, input ProjectCreateInput) (*ent.Project, error) {
	return s.db.Project.Create().
		SetUserID(input.UserID).
		SetName(input.Name).
		SetDescription(input.Description).
		SetStatus("active").
		Save(ctx)
}

// Get returns a project by ID, enforcing ownership.
func (s *ProjectService) Get(ctx context.Context, id string, userID int64) (*ent.Project, error) {
	n, err := parseInt(id)
	if err != nil {
		return nil, errors.New(errors.ErrInvalidRequest, "invalid project id")
	}
	p, err := s.db.Project.Query().
		Where(project.ID(n)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrNotFound, "project not found")
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to fetch project", err)
	}
	if p.UserID != userID {
		return nil, errors.New(errors.ErrForbidden, "access denied")
	}
	return p, nil
}

// List returns all projects for a user.
func (s *ProjectService) List(ctx context.Context, userID int64) ([]*ent.Project, error) {
	return s.db.Project.Query().
		Where(project.UserID(userID)).
		Order(ent.Desc(project.FieldCreatedAt)).
		All(ctx)
}
func parseInt(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

