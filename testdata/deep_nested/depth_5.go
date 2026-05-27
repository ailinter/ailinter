package main

// DeepNested has 5 levels of nesting — should be alert.
func DeepNested(data *Data) error {
	if data != nil {
		if data.IsActive() {
			for _, item := range data.Items {
				if item.IsValid() {
					if item.NeedsProcessing() {
						processItem(item)
					}
				}
			}
		}
	}
	return nil
}
