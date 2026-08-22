package service

import (
	"context"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/collection"
	"github.com/imageforge/imageforge/ent/asset"
	"github.com/imageforge/imageforge/internal/pkg/errors"
)

// CollectionService handles collection business logic.
type CollectionService struct {
	db *ent.Client
}

// NewCollectionService builds a CollectionService.
func NewCollectionService(db *ent.Client) *CollectionService {
	return &CollectionService{db: db}
}

// CollectionCreateInput is the input for creating a collection.
type CollectionCreateInput struct {
	UserID      int64
	Name        string
	Description string
}

// CollectionUpdateInput is the input for updating a collection.
type CollectionUpdateInput struct {
	Name        string
	Description string
}

// Create creates a new collection.
func (s *CollectionService) Create(ctx context.Context, input CollectionCreateInput) (*ent.Collection, error) {
	builder := s.db.Collection.Create().
		SetUserID(input.UserID).
		SetName(input.Name)
	if input.Description != "" {
		builder = builder.SetDescription(input.Description)
	}
	c, err := builder.Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create collection", err)
	}
	return c, nil
}

// Get returns a collection by ID, enforcing ownership.
func (s *CollectionService) Get(ctx context.Context, collectionID, userID int64) (*ent.Collection, error) {
	c, err := s.db.Collection.Query().
		Where(collection.ID(collectionID), collection.UserID(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrNotFound, "collection not found")
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to fetch collection", err)
	}
	return c, nil
}

// List returns all collections for a user.
func (s *CollectionService) List(ctx context.Context, userID int64) ([]*ent.Collection, error) {
	cols, err := s.db.Collection.Query().
		Where(collection.UserID(userID)).
		Order(ent.Desc(collection.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list collections", err)
	}
	return cols, nil
}

// Update modifies a collection's name/description.
func (s *CollectionService) Update(ctx context.Context, collectionID, userID int64, input CollectionUpdateInput) (*ent.Collection, error) {
	c, err := s.Get(ctx, collectionID, userID)
	if err != nil {
		return nil, err
	}
	upd := c.Update().SetName(input.Name)
	if input.Description != "" {
		upd = upd.SetDescription(input.Description)
	} else {
		upd = upd.ClearDescription()
	}
	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to update collection", err)
	}
	return updated, nil
}

// Delete removes a collection.
func (s *CollectionService) Delete(ctx context.Context, collectionID, userID int64) error {
	n, err := s.db.Collection.Delete().
		Where(collection.ID(collectionID), collection.UserID(userID)).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to delete collection", err)
	}
	if n == 0 {
		return errors.New(errors.ErrNotFound, "collection not found")
	}
	return nil
}

// AddAsset adds an asset to a collection (M2M via edge).
func (s *CollectionService) AddAsset(ctx context.Context, collectionID, assetID, userID int64) error {
	c, err := s.Get(ctx, collectionID, userID)
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
	err = c.Update().AddAssetIDs(assetID).Exec(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to add asset to collection", err)
	}
	return nil
}

// RemoveAsset removes an asset from a collection.
func (s *CollectionService) RemoveAsset(ctx context.Context, collectionID, assetID, userID int64) error {
	c, err := s.Get(ctx, collectionID, userID)
	if err != nil {
		return err
	}
	err = c.Update().RemoveAssetIDs(assetID).Exec(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to remove asset from collection", err)
	}
	return nil
}

// ListAssets returns all assets in a collection.
func (s *CollectionService) ListAssets(ctx context.Context, collectionID, userID int64) ([]*ent.Asset, error) {
	c, err := s.Get(ctx, collectionID, userID)
	if err != nil {
		return nil, err
	}
	assets, err := c.QueryAssets().All(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list collection assets", err)
	}
	return assets, nil
}
