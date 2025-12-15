package beam

import (
	"fmt"
	"math"

	"github.com/alexiusacademia/gorcb/internal/nscp"
)

// SinglyReinforced represents a singly reinforced rectangular beam section
type SinglyReinforced struct {
	// Geometry (mm)
	Width         float64 // b - beam width
	Height        float64 // h - total depth
	EffectiveDepth float64 // d - effective depth (to centroid of tension steel)
	Cover         float64 // concrete cover to centroid of reinforcement

	// Materials (MPa)
	Fc float64 // f'c - concrete compressive strength
	Fy float64 // fy - steel yield strength

	// Loading (kN-m)
	Mu float64 // Factored moment

	// Reinforcement (mm²)
	As float64 // Area of tension reinforcement
}

// NewSinglyReinforced creates a new singly reinforced beam with calculated effective depth
func NewSinglyReinforced(width, height, cover, fc, fy float64) *SinglyReinforced {
	return &SinglyReinforced{
		Width:          width,
		Height:         height,
		Cover:          cover,
		EffectiveDepth: height - cover,
		Fc:             fc,
		Fy:             fy,
	}
}

// DesignResult holds the results of beam design
type DesignResult struct {
	// Reinforcement
	AsRequired float64 // Required steel area (mm²)
	AsMin      float64 // Minimum steel area (mm²)
	AsMax      float64 // Maximum steel area (mm²)
	AsProvided float64 // Provided steel area (mm²)

	// Reinforcement ratios
	RhoRequired float64
	RhoMin      float64
	RhoMax      float64
	RhoBalanced float64

	// Section properties
	A          float64 // Depth of compression block (mm)
	C          float64 // Neutral axis depth (mm)
	EpsilonT   float64 // Tensile strain
	Phi        float64 // Strength reduction factor

	// Capacity
	PhiMn float64 // Design moment capacity (kN-m)

	// Status
	IsTensionControlled bool
	IsAdequate          bool
	Message             string
}

// Design calculates the required reinforcement for a given factored moment
func (b *SinglyReinforced) Design(mu float64) (*DesignResult, error) {
	b.Mu = mu

	if b.Width <= 0 || b.EffectiveDepth <= 0 {
		return nil, fmt.Errorf("invalid beam dimensions: width=%.2f, d=%.2f", b.Width, b.EffectiveDepth)
	}
	if b.Fc <= 0 || b.Fy <= 0 {
		return nil, fmt.Errorf("invalid material properties: f'c=%.2f, fy=%.2f", b.Fc, b.Fy)
	}

	result := &DesignResult{}

	// Calculate reinforcement ratio limits
	result.RhoMin = nscp.RhoMin(b.Fc, b.Fy)
	result.RhoMax = nscp.RhoMax(b.Fc, b.Fy)
	result.RhoBalanced = nscp.RhoBalanced(b.Fc, b.Fy)

	// Calculate min and max steel areas
	result.AsMin = result.RhoMin * b.Width * b.EffectiveDepth
	result.AsMax = result.RhoMax * b.Width * b.EffectiveDepth

	// Convert Mu from kN-m to N-mm
	muNmm := mu * 1e6

	// Check if section is adequate for singly reinforced design
	// Maximum moment capacity with tension-controlled section
	beta1 := nscp.Beta1(b.Fc)
	aMax := result.RhoMax * b.Fy * b.Width * b.EffectiveDepth / (0.85 * b.Fc * b.Width)
	phiMnMax := nscp.PhiFlexure * 0.85 * b.Fc * b.Width * aMax * (b.EffectiveDepth - aMax/2) / 1e6

	if mu > phiMnMax {
		result.IsAdequate = false
		result.Message = fmt.Sprintf("Section inadequate for singly reinforced design. Mu=%.2f kN-m > φMn,max=%.2f kN-m. Consider increasing section size or using doubly reinforced design.", mu, phiMnMax)
		result.PhiMn = phiMnMax
		return result, nil
	}

	// Calculate required steel using iterative approach
	// Start with assuming φ = 0.90 (tension-controlled)
	phi := nscp.PhiFlexure

	// Rn = Mu / (φ * b * d²)
	Rn := muNmm / (phi * b.Width * math.Pow(b.EffectiveDepth, 2))

	// ρ = (0.85*f'c/fy) * (1 - √(1 - 2*Rn/(0.85*f'c)))
	term := 2 * Rn / (0.85 * b.Fc)
	if term > 1 {
		result.IsAdequate = false
		result.Message = "Section inadequate - moment too high for singly reinforced design"
		return result, nil
	}

	rhoRequired := (0.85 * b.Fc / b.Fy) * (1 - math.Sqrt(1-term))
	result.RhoRequired = rhoRequired

	// Check against minimum
	if rhoRequired < result.RhoMin {
		rhoRequired = result.RhoMin
	}

	result.AsRequired = rhoRequired * b.Width * b.EffectiveDepth

	// Verify the design - calculate actual a and c
	result.A = result.AsRequired * b.Fy / (0.85 * b.Fc * b.Width)
	result.C = result.A / beta1

	// Calculate tensile strain
	result.EpsilonT = nscp.EpsilonCU * (b.EffectiveDepth - result.C) / result.C

	// Recalculate phi based on actual strain
	result.Phi = nscp.Phi(result.EpsilonT, b.Fy)
	result.IsTensionControlled = result.EpsilonT >= 0.005

	// Calculate actual capacity
	result.PhiMn = result.Phi * result.AsRequired * b.Fy * (b.EffectiveDepth - result.A/2) / 1e6

	result.IsAdequate = result.PhiMn >= mu
	result.AsProvided = result.AsRequired

	if result.IsAdequate {
		result.Message = "Design OK - Section is tension-controlled"
		if !result.IsTensionControlled {
			result.Message = "Design OK - Section is in transition zone"
		}
	}

	return result, nil
}

// AnalysisResult holds the results of beam analysis
type AnalysisResult struct {
	// Section properties
	A        float64 // Depth of compression block (mm)
	C        float64 // Neutral axis depth (mm)
	Beta1    float64 // Stress block factor
	EpsilonT float64 // Tensile strain
	Phi      float64 // Strength reduction factor

	// Reinforcement ratios
	Rho         float64
	RhoMin      float64
	RhoMax      float64
	RhoBalanced float64

	// Capacity
	Mn    float64 // Nominal moment capacity (kN-m)
	PhiMn float64 // Design moment capacity (kN-m)

	// Status
	IsTensionControlled bool
	MeetsMinReinf       bool
	MeetsMaxReinf       bool
	Message             string
}

// Analyze calculates the moment capacity for a given reinforcement area
func (b *SinglyReinforced) Analyze(as float64) (*AnalysisResult, error) {
	b.As = as

	if b.Width <= 0 || b.EffectiveDepth <= 0 {
		return nil, fmt.Errorf("invalid beam dimensions: width=%.2f, d=%.2f", b.Width, b.EffectiveDepth)
	}
	if b.Fc <= 0 || b.Fy <= 0 {
		return nil, fmt.Errorf("invalid material properties: f'c=%.2f, fy=%.2f", b.Fc, b.Fy)
	}
	if as <= 0 {
		return nil, fmt.Errorf("invalid reinforcement area: As=%.2f", as)
	}

	result := &AnalysisResult{}
	result.Beta1 = nscp.Beta1(b.Fc)

	// Calculate reinforcement ratio limits
	result.RhoMin = nscp.RhoMin(b.Fc, b.Fy)
	result.RhoMax = nscp.RhoMax(b.Fc, b.Fy)
	result.RhoBalanced = nscp.RhoBalanced(b.Fc, b.Fy)

	// Actual reinforcement ratio
	result.Rho = as / (b.Width * b.EffectiveDepth)

	// Check min/max reinforcement
	result.MeetsMinReinf = result.Rho >= result.RhoMin
	result.MeetsMaxReinf = result.Rho <= result.RhoMax

	// Calculate depth of compression block
	// T = C → As*fy = 0.85*f'c*b*a
	result.A = as * b.Fy / (0.85 * b.Fc * b.Width)
	result.C = result.A / result.Beta1

	// Calculate tensile strain
	result.EpsilonT = nscp.EpsilonCU * (b.EffectiveDepth - result.C) / result.C

	// Determine phi based on strain
	result.Phi = nscp.Phi(result.EpsilonT, b.Fy)
	result.IsTensionControlled = result.EpsilonT >= 0.005

	// Calculate moment capacity
	// Mn = As * fy * (d - a/2)
	result.Mn = as * b.Fy * (b.EffectiveDepth - result.A/2) / 1e6
	result.PhiMn = result.Phi * result.Mn

	// Build status message
	if result.IsTensionControlled {
		result.Message = "Section is tension-controlled (εt ≥ 0.005)"
	} else if result.EpsilonT >= b.Fy/nscp.Es {
		result.Message = "Section is in transition zone"
	} else {
		result.Message = "Section is compression-controlled (εt < εy)"
	}

	if !result.MeetsMinReinf {
		result.Message += " | WARNING: Below minimum reinforcement"
	}
	if !result.MeetsMaxReinf {
		result.Message += " | WARNING: Exceeds maximum reinforcement"
	}

	return result, nil
}

