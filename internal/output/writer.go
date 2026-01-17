package output

import "github.com/ervinmplayon/tractatus/internal/inventory"

type OutputWriter interface {
	Write(inv *inventory.Inventory) error
}
