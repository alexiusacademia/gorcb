package section

import "fmt"

// Section represents a non-rectangular concrete section defined by vertices
// The section is defined in a local coordinate system where:
// - Y-axis points upward (positive = compression zone at top)
// - X-axis points to the right
// - Origin can be at any convenient location
type Section struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	
	// Material properties
	Fc float64 `json:"fc"` // Concrete compressive strength (MPa)
	Fy float64 `json:"fy"` // Steel yield strength (MPa)

	// Section geometry defined by vertices (in mm)
	// Vertices should be defined counter-clockwise for the outer boundary
	// The section is assumed to be a simple polygon (no holes)
	Vertices []Point `json:"vertices"`

	// Reinforcement layers
	Reinforcement []RebarLayer `json:"reinforcement"`

	// Effective depth override (optional, calculated from reinforcement if not provided)
	EffectiveDepth float64 `json:"effective_depth,omitempty"`
}

// Point represents a 2D coordinate
type Point struct {
	X float64 `json:"x"` // mm
	Y float64 `json:"y"` // mm
}

// RebarLayer represents a layer of reinforcement at a specific depth
type RebarLayer struct {
	// Position of the reinforcement layer centroid
	Y float64 `json:"y"` // mm from bottom of section

	// Reinforcement area in this layer
	Area float64 `json:"area"` // mm²

	// Optional: description of bars (e.g., "3-25mm")
	Description string `json:"description,omitempty"`

	// Type: "tension" or "compression" (default: auto-detect based on position)
	Type string `json:"type,omitempty"`
}

// SectionProperties holds calculated geometric properties
type SectionProperties struct {
	// Overall dimensions
	Width      float64 // Maximum width (mm)
	Height     float64 // Total height (mm)
	Area       float64 // Gross area (mm²)
	
	// Centroid location
	CentroidX float64 // mm
	CentroidY float64 // mm

	// Bounding box
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64

	// Reinforcement summary
	TotalTensionSteel     float64 // mm²
	TotalCompressionSteel float64 // mm²
	EffectiveDepth        float64 // mm (to centroid of tension steel)
	CompressionCover      float64 // mm (to centroid of compression steel)
}

// Validate checks if the section definition is valid
func (s *Section) Validate() error {
	if len(s.Vertices) < 3 {
		return &ValidationError{"section must have at least 3 vertices"}
	}
	if s.Fc <= 0 {
		return &ValidationError{"f'c must be positive"}
	}
	if s.Fy <= 0 {
		return &ValidationError{"fy must be positive"}
	}
	if len(s.Reinforcement) == 0 {
		return &ValidationError{"section must have at least one reinforcement layer"}
	}
	for i, layer := range s.Reinforcement {
		if layer.Area <= 0 {
			return &ValidationError{msg: fmt.Sprintf("reinforcement layer %d must have positive area", i+1)}
		}
	}
	return nil
}

// ValidationError represents a section validation error
type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return e.msg
}

