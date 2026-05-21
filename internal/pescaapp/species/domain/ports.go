package domain

import "context"

type SpeciesRepository interface {
	FindByID(ctx context.Context, id string) (*Species, error)
	List(ctx context.Context, limit, offset int, search string) ([]*Species, int, error)
	Create(ctx context.Context, species *Species) (*Species, error)
	Update(ctx context.Context, id string, req UpdateSpeciesRequest) (*Species, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit int) ([]*Species, error)
}

