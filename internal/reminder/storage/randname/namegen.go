package randname

import (
	"fmt"
	"math/rand"
)

var (
	adjectives = [...]string{
		"Timely", "Fresh", "Preserving", "Seasonal", "Expiring", "Rotating",
		"Temporary", "Chilled", "Regulated", "Cool", "Prompt", "Cyclical",
		"Periodic", "Interval", "Short-term", "Long-term", "Dated", "Turnover",
		"Shifting", "Revolving",
	}
	nouns = [...]string{
		"Storage", "Depot", "Refrigerator", "Freezer", "Cellar", "Pantry",
		"Cabinet", "Shelf", "Bin", "Locker", "Container", "Cubby", "Vault",
		"Closet", "Rack", "Unit", "Warehouse", "Silo", "Cooler", "Icebox",
	}
)

func StorageName() string {
	var (
		adjective = adjectives[rand.Intn(len(adjectives))]
		noun      = nouns[rand.Intn(len(nouns))]
	)

	return adjective + noun
}

func StorageNameWithDigits() string {
	const maxNumber = 1e5

	var (
		adjective = adjectives[rand.Intn(len(adjectives))]
		noun      = nouns[rand.Intn(len(nouns))]
		digits    = rand.Intn(maxNumber)
	)

	return fmt.Sprintf("%s%s%d", adjective, noun, digits)
}
