package main

// ProcessOrder has 3 levels of nesting.
func ProcessOrder(order *Order) error {
	if order != nil {
		if order.IsValid() {
			if order.HasItems() {
				return shipOrder(order)
			}
		}
	}
	return nil
}
