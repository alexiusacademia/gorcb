package section

import (
	"encoding/json"
	"math"
	"os"

	"github.com/alexiusacademia/gorcb/internal/nscp"
)

// LoadFromFile loads a section definition from a JSON file
func LoadFromFile(filepath string) (*Section, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var section Section
	if err := json.Unmarshal(data, &section); err != nil {
		return nil, err
	}

	if err := section.Validate(); err != nil {
		return nil, err
	}

	return &section, nil
}

// AnalysisResult holds the results of section analysis
type AnalysisResult struct {
	// Section properties
	Properties *SectionProperties

	// Neutral axis and compression block
	C     float64 // Neutral axis depth from top (mm)
	A     float64 // Compression block depth (mm)
	Beta1 float64 // Stress block factor

	// Compression zone
	CompressionArea     float64 // Area of compression block (mm²)
	CompressionCentroid float64 // Depth to centroid of compression block (mm)

	// Strains
	EpsilonT float64 // Maximum tensile strain (at lowest tension steel)

	// Forces (kN)
	Cc float64 // Concrete compression force
	Cs float64 // Compression steel force (if any)
	T  float64 // Total tension steel force

	// Steel layer details
	SteelLayers []SteelLayerResult

	// Capacity
	Phi   float64 // Strength reduction factor
	Mn    float64 // Nominal moment capacity (kN-m)
	PhiMn float64 // Design moment capacity (kN-m)

	// Status
	IsTensionControlled bool
	Message             string
}

// SteelLayerResult holds analysis results for each reinforcement layer
type SteelLayerResult struct {
	Y           float64 // Position from section bottom (mm)
	Area        float64 // Steel area (mm²)
	Strain      float64 // Strain at this layer
	Stress      float64 // Stress (MPa)
	Force       float64 // Force (kN)
	IsTension   bool    // True if in tension
	HasYielded  bool    // True if steel has yielded
	Description string
}

// Analyze calculates the moment capacity of the section
func (s *Section) Analyze() (*AnalysisResult, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	result := &AnalysisResult{}
	result.Properties = s.CalculateProperties()
	result.Beta1 = nscp.Beta1(s.Fc)

	// Find neutral axis by iteration (force equilibrium)
	// T = Cc + Cs
	props := result.Properties
	
	// Initial guess for c: assume tension-controlled
	c := props.EffectiveDepth * 0.3

	epsilonY := s.Fy / nscp.Es

	// Iterate to find neutral axis
	for iter := 0; iter < 100; iter++ {
		a := result.Beta1 * c

		// Calculate concrete compression force
		compArea := s.CompressionBlockArea(a)
		Cc := 0.85 * s.Fc * compArea / 1000 // kN

		// Calculate steel forces
		var totalTension, totalCompression float64
		result.SteelLayers = nil

		for _, layer := range s.Reinforcement {
			// Neutral axis is at depth c from top
			// Layer is at Y from bottom, so from top it's (MaxY - Y)
			depthFromTop := props.MaxY - layer.Y

			// Strain at this layer
			strain := nscp.EpsilonCU * (c - depthFromTop) / c

			// Stress (limited to fy)
			var stress float64
			if strain >= 0 {
				// Compression
				stress = math.Min(strain*nscp.Es, s.Fy)
			} else {
				// Tension
				stress = math.Max(strain*nscp.Es, -s.Fy)
			}

			force := layer.Area * stress / 1000 // kN

			layerResult := SteelLayerResult{
				Y:           layer.Y,
				Area:        layer.Area,
				Strain:      strain,
				Stress:      stress,
				Force:       force,
				IsTension:   strain < 0,
				HasYielded:  math.Abs(strain) >= epsilonY,
				Description: layer.Description,
			}
			result.SteelLayers = append(result.SteelLayers, layerResult)

			if strain >= 0 {
				// Compression steel - subtract displaced concrete if within compression block
				if depthFromTop <= a {
					// Within compression block, subtract displaced concrete
					netStress := stress - 0.85*s.Fc
					force = layer.Area * netStress / 1000
				}
				totalCompression += force
			} else {
				totalTension += math.Abs(force)
			}
		}

		// Check equilibrium: T = Cc + Cs
		totalComp := Cc + totalCompression
		imbalance := totalTension - totalComp

		if math.Abs(imbalance) < 0.1 { // Converged (within 0.1 kN)
			result.C = c
			result.A = a
			result.CompressionArea = compArea
			result.CompressionCentroid = s.CompressionBlockCentroid(a)
			result.Cc = Cc
			result.Cs = totalCompression
			result.T = totalTension
			break
		}

		// Adjust c based on imbalance
		// If T > C, need more compression, so increase c
		// If T < C, need less compression, so decrease c
		adjustment := imbalance / (0.85 * s.Fc * s.WidthAtDepth(c) / 1000)
		adjustment = math.Max(math.Min(adjustment, 10), -10) // Limit adjustment
		c += adjustment * 0.5 // Damped adjustment

		// Keep c within bounds
		c = math.Max(c, 1)
		c = math.Min(c, props.Height-1)
	}

	// Find maximum tensile strain (at bottom-most tension steel)
	var maxTensileStrain float64
	for _, layer := range result.SteelLayers {
		if layer.IsTension && math.Abs(layer.Strain) > math.Abs(maxTensileStrain) {
			maxTensileStrain = layer.Strain
		}
	}
	result.EpsilonT = math.Abs(maxTensileStrain)

	// Determine phi
	result.Phi = nscp.Phi(result.EpsilonT, s.Fy)
	result.IsTensionControlled = result.EpsilonT >= 0.005

	// Calculate moment capacity about the top of section
	// Then convert to about the tension steel centroid
	// Mn = Cc * (d - ȳc) + Σ(steel moments)
	
	d := props.EffectiveDepth
	Mn := result.Cc * (d - result.CompressionCentroid)

	// Add compression steel contribution
	for _, layer := range result.SteelLayers {
		if !layer.IsTension {
			depthFromTop := props.MaxY - layer.Y
			Mn += layer.Force * (d - depthFromTop)
		}
	}

	result.Mn = Mn / 1000 // Convert to kN-m
	result.PhiMn = result.Phi * result.Mn

	// Status message
	if result.IsTensionControlled {
		result.Message = "Section is tension-controlled (εt ≥ 0.005)"
	} else if result.EpsilonT >= epsilonY {
		result.Message = "Section is in transition zone"
	} else {
		result.Message = "Section is compression-controlled"
	}

	return result, nil
}

// DesignResult holds the results of section design
type DesignResult struct {
	// Input
	Mu float64 // Required moment (kN-m)

	// Section properties
	Properties *SectionProperties

	// Design results
	AsRequired float64 // Required tension steel area (mm²)
	AsMin      float64 // Minimum steel area (mm²)
	AsProvided float64 // Provided steel area (mm²)

	// Section at capacity
	C     float64 // Neutral axis depth (mm)
	A     float64 // Compression block depth (mm)
	Beta1 float64

	// Capacity check
	Phi   float64
	PhiMn float64 // Achieved capacity (kN-m)

	// Status
	IsTensionControlled bool
	IsAdequate          bool
	Message             string
}

// Design calculates the required reinforcement for a given moment
// Note: For non-rectangular sections, this modifies the tension steel area
// while keeping compression steel (if any) as defined
func (s *Section) Design(mu float64) (*DesignResult, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	result := &DesignResult{
		Mu: mu,
	}
	result.Properties = s.CalculateProperties()
	result.Beta1 = nscp.Beta1(s.Fc)

	props := result.Properties
	d := props.EffectiveDepth

	// Calculate minimum steel area
	result.AsMin = nscp.RhoMin(s.Fc, s.Fy) * props.Width * d

	// Iterative design: adjust tension steel until capacity matches demand
	// Start with an estimate based on rectangular section formula
	phi := nscp.PhiFlexure
	muNmm := mu * 1e6

	// Estimate lever arm as 0.9d
	jd := 0.9 * d
	AsEstimate := muNmm / (phi * s.Fy * jd)

	// Create a working copy of the section to modify reinforcement
	workingSection := *s

	// Find or create the tension steel layer
	tensionLayerIdx := -1
	for i, layer := range workingSection.Reinforcement {
		if layer.Type == "tension" || (layer.Type == "" && layer.Y < props.Height/2) {
			tensionLayerIdx = i
			break
		}
	}

	if tensionLayerIdx < 0 {
		// Use the bottom-most layer as tension
		minY := workingSection.Reinforcement[0].Y
		tensionLayerIdx = 0
		for i, layer := range workingSection.Reinforcement {
			if layer.Y < minY {
				minY = layer.Y
				tensionLayerIdx = i
			}
		}
	}

	// Iterate to find required As
	As := AsEstimate
	for iter := 0; iter < 50; iter++ {
		workingSection.Reinforcement[tensionLayerIdx].Area = As

		analysis, err := workingSection.Analyze()
		if err != nil {
			return nil, err
		}

		if analysis.PhiMn >= mu*0.999 {
			// Adequate
			result.AsRequired = As
			result.C = analysis.C
			result.A = analysis.A
			result.Phi = analysis.Phi
			result.PhiMn = analysis.PhiMn
			result.IsTensionControlled = analysis.IsTensionControlled
			result.IsAdequate = true
			break
		}

		// Increase As proportionally
		ratio := mu / analysis.PhiMn
		As *= ratio
		As = math.Min(As, props.Width*d*nscp.RhoMax(s.Fc, s.Fy)*3) // Limit to prevent infinite loop
	}

	// Check against minimum
	if result.AsRequired < result.AsMin {
		result.AsRequired = result.AsMin
	}

	result.AsProvided = result.AsRequired

	// Build message
	if result.IsAdequate {
		if result.IsTensionControlled {
			result.Message = "Design OK - Section is tension-controlled"
		} else {
			result.Message = "Design OK - Section is in transition zone"
		}
	} else {
		result.Message = "Design inadequate - Section cannot resist the required moment"
	}

	return result, nil
}

