package service

import (
	"context"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/favorite"
	"github.com/imageforge/imageforge/ent/asset"
	"github.com/imageforge/imageforge/internal/pkg/errors"
)

// FavoriteService handles favorite business logic.
type FavoriteService struct {
	db *ent.Client
}

// NewFavoriteService builds a FavoriteService.
func NewFavoriteService(db *ent.Client) *FavoriteService {
	return &FavoriteService{db: db}
}

// Add creates a favorite record for a user+asset pair.
func (s *FavoriteService) Add(ctx context.Context, userID, assetID int64) (*ent.Favorite, error) {
	// Verify asset exists and belongs to the user.
	_, err := s.db.Asset.Query().
		Where(asset.ID(assetID), asset.UserID(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrNotFound, "asset not found")
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to verify asset", err)
	}

	// Idempotent: already favorited is not an error.
	existing, err := s.db.Favorite.Query().
		Where(favorite.UserID(userID), favorite.AssetID(assetID)).
		Only(ctx)
	if err == nil {
		return existing, nil
	}
	if !ent.IsNotFound(err) {
		return nil, errors.Wrap(errors.ErrInternal, "failed to check favorite", err)
	}

	f, err := s.db.Favorite.Create().
		SetUserID(userID).
		SetAssetID(assetID).
		Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to add favorite", err)
	}
	return f, nil
}

// Remove deletes a favorite record for a user+asset pair.
func (s *FavoriteService) Remove(ctx context.Context, userID, assetID int64) error {
	n, err := s.db.Favorite.Delete().
		Where(favorite.UserID(userID), favorite.AssetID(assetID)).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to remove favorite", err)
	}
	if n == 0 {
		return errors.New(errors.ErrNotFound, "favorite not found")
	}
	return nil
}

// List returns all favorites for a user, newest first.
func (s *FavoriteService) List(ctx context.Context, userID int64) ([]*ent.Favorite, error) {
	favs, err := s.db.Favorite.Query().
		Where(favorite.UserID(userID)).
		Order(ent.Desc(favorite.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list favorites", err)
	}
	return favs, nil
}

// IsFavorited checks whether a user has favorited a specific asset.
func (s *FavoriteService) IsFavorited(ctx context.Context, userID, assetID int64) (bool, error) {
	exists, err := s.db.Favorite.Query().
		Where(favorite.UserID(userID), favorite.AssetID(assetID)).
		Exist(ctx)
	if err != nil {
		return false, errors.Wrap(errors.ErrInternal, "failed to check favorite", err)
	}
	return exists, nil
}

// ListFavoritedAssets returns asset IDs favorited by the user.
func (s *FavoriteService) ListFavoritedAssets(ctx context.Context, userID int64) ([]int64, error) {
	favs, err := s.db.Favorite.Query().
		Where(favorite.UserID(userID)).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list favorite assets", err)
	}
	ids := make([]int64, len(favs))
	for i, f := range favs {
		ids[i] = f.AssetID
	}
	return ids, nil
}
