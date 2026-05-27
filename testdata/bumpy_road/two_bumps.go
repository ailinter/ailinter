package main

// ProcessReport has 2 bumps of nested logic.
func ProcessReport(data []Record) *Report {
	r := &Report{}

	// Bump 1: Validation
	for _, d := range data {
		if d.IsActive {
			if d.HasRequiredFields() {
				validateRecord(d, r)
			}
		}
	}

	// Bump 2: Aggregation
	for _, d := range data {
		if d.Type == "sale" {
			if d.Amount > 0 {
				aggregateSale(d, r)
			}
		}
	}

	return r
}
