package nscp

// LoadCombination represents an NSCP load combination
// Based on NSCP 2015 Section 203.3 - Load Combinations Using Strength Design
type LoadCombination struct {
	ID          string
	Description string
	// Load factors for each load type
	Dead      float64 // D - Dead load
	Live      float64 // L - Live load
	Roof      float64 // Lr - Roof live load
	Wind      float64 // W - Wind load
	Earthquake float64 // E - Earthquake load
	Rain      float64 // R - Rain load
}

// NSCP 2015 Section 203.3.1 - Basic Load Combinations
var LoadCombinations = []LoadCombination{
	{
		ID:          "1",
		Description: "1.4D",
		Dead:        1.4,
	},
	{
		ID:          "2",
		Description: "1.2D + 1.6L + 0.5(Lr or R)",
		Dead:        1.2,
		Live:        1.6,
		Roof:        0.5,
		Rain:        0.5,
	},
	{
		ID:          "3",
		Description: "1.2D + 1.6(Lr or R) + (1.0L or 0.5W)",
		Dead:        1.2,
		Live:        1.0,
		Roof:        1.6,
		Rain:        1.6,
		Wind:        0.5,
	},
	{
		ID:          "4",
		Description: "1.2D + 1.0W + 1.0L + 0.5(Lr or R)",
		Dead:        1.2,
		Live:        1.0,
		Wind:        1.0,
		Roof:        0.5,
		Rain:        0.5,
	},
	{
		ID:          "5",
		Description: "1.2D + 1.0E + 1.0L",
		Dead:        1.2,
		Live:        1.0,
		Earthquake:  1.0,
	},
	{
		ID:          "6",
		Description: "0.9D + 1.0W",
		Dead:        0.9,
		Wind:        1.0,
	},
	{
		ID:          "7",
		Description: "0.9D + 1.0E",
		Dead:        0.9,
		Earthquake:  1.0,
	},
}

// SimplifiedCombinations for common beam design scenarios
// These are the most frequently used combinations for gravity loads
var SimplifiedCombinations = []LoadCombination{
	{
		ID:          "1",
		Description: "1.4D",
		Dead:        1.4,
	},
	{
		ID:          "2",
		Description: "1.2D + 1.6L",
		Dead:        1.2,
		Live:        1.6,
	},
}

// CalculateFactoredMoment calculates the factored moment for a given load combination
func (lc LoadCombination) CalculateFactoredMoment(moments LoadMoments) float64 {
	return lc.Dead*moments.Dead +
		lc.Live*moments.Live +
		lc.Roof*moments.Roof +
		lc.Wind*moments.Wind +
		lc.Earthquake*moments.Earthquake +
		lc.Rain*moments.Rain
}

// LoadMoments holds unfactored moments from different load types
type LoadMoments struct {
	Dead       float64 // Moment due to dead load (kN-m)
	Live       float64 // Moment due to live load (kN-m)
	Roof       float64 // Moment due to roof live load (kN-m)
	Wind       float64 // Moment due to wind load (kN-m)
	Earthquake float64 // Moment due to earthquake load (kN-m)
	Rain       float64 // Moment due to rain load (kN-m)
}

// CalculateGoverningMoment finds the maximum factored moment from all combinations
func CalculateGoverningMoment(moments LoadMoments, combinations []LoadCombination) (float64, LoadCombination) {
	var maxMoment float64
	var governingCombo LoadCombination

	for _, combo := range combinations {
		mu := combo.CalculateFactoredMoment(moments)
		if mu > maxMoment {
			maxMoment = mu
			governingCombo = combo
		}
	}

	return maxMoment, governingCombo
}

