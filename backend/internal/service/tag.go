package service

import (
	"context"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/tag"
	"github.com/imageforge/imageforge/ent/assettag"
	"github.com/imageforge/imageforge/ent/asset"
	"github.com/imageforge/imageforge/internal/pkg/errors"
)

// TagService handles tag business logic.
type TagService struct {
	db *ent.Client
}

// NewTagService builds a TagService.
func NewTagService(db *ent.Client) *TagService {
	return &TagService{db: db}
}

// TagCreateInput is the input for creating a tag.
type TagCreateInput struct {
	UserID int64
	Name   string
	Color  string
}

// TagUpdateInput is the input for updating a tag.
type TagUpdateInput struct {
	Name  string
	Color string
}

// Create creates a new tag.
func (s *TagService) Create(ctx context.Context, input TagCreateInput) (*ent.Tag, error) {
	builder := s.db.Tag.Create().
		SetUserID(input.UserID).
		SetName(input.Name)
	if input.Color != "" {
		builder = builder.SetColor(input.Color)
	}
	t, err := builder.Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create tag", err)
	}
	return t, nil
}

// Get returns a tag by ID, enforcing ownership.
func (s *TagService) Get(ctx context.Context, tagID, userID int64) (*ent.Tag, error) {
	t, err := s.db.Tag.Query().
		Where(tag.ID(tagID), tag.UserID(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrNotFound, "tag not found")
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to fetch tag", err)
	}
	return t, nil
}

// List returns all tags for a user.
func (s *TagService) List(ctx context.Context, userID int64) ([]*ent.Tag, error) {
	tags, err := s.db.Tag.Query().
		Where(tag.UserID(userID)).
		Order(ent.Asc(tag.FieldName)).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list tags", err)
	}
	return tags, nil
}

// Update modifies a tag's name/color.
func (s *TagService) Update(ctx context.Context, tagID, userID int64, input TagUpdateInput) (*ent.Tag, error) {
	t, err := s.Get(ctx, tagID, userID)
	if err != nil {
		return nil, err
	}
	upd := t.Update().SetName(input.Name)
	if input.Color != "" {
		upd = upd.SetColor(input.Color)
	} else {
		upd = upd.ClearColor()
	}
	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to update tag", err)
	}
	return updated, nil
}

// Delete removes a tag.
func (s *TagService) Delete(ctx context.Context, tagID, userID int64) error {
	n, err := s.db.Tag.Delete().
		Where(tag.ID(tagID), tag.UserID(userID)).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to delete tag", err)
	}
	if n == 0 {
		return errors.New(errors.ErrNotFound, "tag not found")
	}
	return nil
}

// TagAsset links an asset to a tag (M2M).
func (s *TagService) TagAsset(ctx context.Context, tagID, assetID, userID int64) error {
	t, err := s.Get(ctx, tagID, userID)
	if err != nil {
		return err
	}
	// Verify asset exists and belongs to user.
	_, err = s.db.Asset.Query().
		Where(asset.ID(assetID), asset.UserID(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New(errors.ErrNotFound, "asset not found")
		}
		return errors.Wrap(errors.ErrInternal, "failed to verify asset", err)
	}
	// Idempotent: skip if already tagged.
	exists, err := s.db.AssetTag.Query().
		Where(assettag.AssetID(assetID), assettag.TagID(tagID)).
		Exist(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to check tag", err)
	}
	if exists {
		return nil
	}
	_, err = s.db.AssetTag.Create().
		SetAssetID(assetID).
		SetTagID(tagID).
		Save(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to tag asset", err)
	}
	// Touch tag update time.
	_ = t.Update().Exec(ctx)
	return nil
}

// UntagAsset removes the link between an asset and a tag.
func (s *TagService) UntagAsset(ctx context.Context, tagID, assetID, userID int64) error {
	// Verify tag belongs to user.
	_, err := s.Get(ctx, tagID, userID)
	if err != nil {
		return err
	}
	n, err := s.db.AssetTag.Delete().
		Where(assettag.AssetID(assetID), assettag.TagID(tagID)).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to untag asset", err)
	}
	if n == 0 {
		return errors.New(errors.ErrNotFound, "asset-tag link not found")
	}
	return nil
}

// ListAssets returns all assets carrying a specific tag.
func (s *TagService) ListAssets(ctx context.Context, tagID, userID int64) ([]*ent.Asset, error) {
	t, err := s.Get(ctx, tagID, userID)
	if err != nil {
		return nil, err
	}
	assets, err := t.QueryAssets().All(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list tagged assets", err)
	}
	return assets, nil
}

// ListTagsForAsset returns all tags applied to an asset by a user.
func (s *TagService) ListTagsForAsset(ctx context.Context, assetID, userID int64) ([]*ent.Tag, error) {
	// Verify asset belongs to user.
	_, err := s.db.Asset.Query().
		Where(asset.ID(assetID), asset.UserID(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrNotFound, "asset not found")
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to verify asset", err)
	}
	tags, err := s.db.Tag.Query().
		Where(tag.UserID(userID), tag.HasAssetTagsWith(assettag.AssetID(assetID))).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list asset tags", err)
	}
	return tags, nil
}
