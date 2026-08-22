package service

import (
	"context"
	"fmt"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/asset"
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

// AssetService handles asset business logic.
type AssetService struct {
	db *ent.Client
}

// NewAssetService builds an AssetService.
func NewAssetService(db *ent.Client) *AssetService {
	return &AssetService{db: db}
}

// AssetFilter is the filter for listing assets.
type AssetFilter struct {
	ProjectID string
	Tag       string
	Page      int
	PageSize  int
}

// List returns assets for a user with filtering and pagination.
func (s *AssetService) List(ctx context.Context, userID int64, filter AssetFilter) ([]*ent.Asset, int, error) {
	query := s.db.Asset.Query().
		Where(asset.UserID(userID))

	if filter.ProjectID != "" {
		pid, err := parseInt(filter.ProjectID)
		if err == nil {
			query = query.Where(asset.ProjectID(pid))
		}
	}

	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(errors.ErrInternal, "failed to count assets", err)
	}

	assets, err := query.
		Order(ent.Desc(asset.FieldCreatedAt)).
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(errors.ErrInternal, "failed to list assets", err)
	}
	return assets, total, nil
}

// Get returns an asset by ID, enforcing ownership.
func (s *AssetService) Get(ctx context.Context, id int64, userID int64) (*ent.Asset, error) {
	a, err := s.db.Asset.Query().
		Where(asset.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrNotFound, "asset not found")
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to fetch asset", err)
	}
	if a.UserID != userID {
		return nil, errors.New(errors.ErrForbidden, "access denied")
	}
	return a, nil
}

// Delete soft-deletes an asset.
func (s *AssetService) Delete(ctx context.Context, id int64, userID int64) error {
	a, err := s.Get(ctx, id, userID)
	if err != nil {
		return err
	}
	_, err = s.db.Asset.UpdateOneID(a.ID).
		SetStatus("deleted").
		Save(ctx)
	return err
}

func parseInt(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// GetContent returns the raw image bytes for an asset.
func (s *AssetService) GetContent(ctx context.Context, id int64, userID int64) ([]byte, string, error) {
	a, err := s.Get(ctx, id, userID)
	if err != nil {
		return nil, "", err
	}
	// In a full implementation, this fetches from object storage using a.StorageKey.
	// For now, return a placeholder.
	return []byte{}, a.MimeType, nil
}
