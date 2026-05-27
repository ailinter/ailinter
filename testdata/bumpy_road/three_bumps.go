package main

func ThreeBumps(items []Item) int {
	count := 0

	// Bump 1
	for _, it := range items {
		if it.IsActive() {
			if it.Quantity > 0 {
				count += it.Quantity
			}
		}
	}

	// Bump 2
	for _, it := range items {
		if it.Type == "premium" {
			if it.Price > 100 {
				count += int(it.Price)
			}
		}
	}

	// Bump 3
	if count > 1000 {
		if items[0].Category == "luxury" {
			count = count * 2
		}
	}

	return count
}
