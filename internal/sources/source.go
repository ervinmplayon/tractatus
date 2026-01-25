package sources

import (
	"context"

	"github.com/ervinmplayon/tractatus/internal/inventory"
)

type DataSource interface {
	Collect(ctx context.Context) ([]*inventory.ResourceInfo, error)
	Name() string
}
