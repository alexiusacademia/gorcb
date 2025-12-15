package beam

import (
	"fmt"
	"math"

	"github.com/alexiusacademia/gorcb/internal/nscp"
)

// DoublyReinforced represents a doubly reinforced rectangular beam section
type DoublyReinforced struct {
	// Geometry (mm)
	Width          float64 // b - beam width
	Height         float64 // h - total depth
	EffectiveDepth float64 // d - effective depth (to centroid of tension steel)
	Cover          float64 // concrete cover to centroid of tension reinforcement
	CoverComp      float64 // d' - cover to centroid of compression reinforcement

	// Materials (MPa)
	Fc float64 // f'c - concrete compressive strength
	Fy float64 // fy - steel yield strength

	// Loading (kN-m)
	Mu float64 // Factored moment

	// Reinforcement (mm²)
	As  float64 // Area of tension reinforcement
	Asc float64 // Area of compression reinforcement
}

// NewDoublyReinforced creates a new doubly reinforced beam
func NewDoublyReinforced(width, height, cover, coverComp, fc, fy float64) *DoublyReinforced {
	return &DoublyReinforced{
		Width:          width,
		Height:         height,
		Cover:          cover,
		CoverComp:      coverComp,
		EffectiveDepth: height - cover,
		Fc:             fc,
		Fy:             fy,
	}
}

// DoublyDesignResult holds the results of doubly reinforced beam design
type DoublyDesignResult struct {
	// Is doubly reinforced needed?
	RequiresCompSteel bool

	// Moment components
	Mu1 float64 // Moment resisted by tension steel with concrete (kN-m)
	Mu2 float64 // Moment resisted by steel couple (kN-m)

	// Reinforcement
	As1        float64 // Tension steel for concrete compression (mm²)
	As2        float64 // Additional tension steel for compression steel (mm²)
	AsTotal    float64 // Total tension reinforcement (mm²)
	AscRequired float64 // Required compression reinforcement (mm²)

	// Limits
	AsMin float64 // Minimum tension steel (mm²)
	AsMax float64 // Maximum tension steel for singly reinforced (mm²)

	// Reinforcement ratios
	RhoMin      float64
	RhoMax      float64
	RhoBalanced float64

	// Section properties at max capacity (singly)
	AMax float64 // Maximum a for tension-controlled (mm)
	CMax float64 // Maximum c for tension-controlled (mm)

	// Compression steel stress
	FscStress    float64 // Actual stress in compression steel (MPa)
	CompYielded  bool    // Whether compression steel has yielded

	// Strains
	EpsilonT  float64 // Tensile strain
	EpsilonSc float64 // Compression steel strain

	// Capacity
	Phi   float64 // Strength reduction factor
	PhiMn float64 // Design moment capacity (kN-m)

	// Status
	IsTensionControlled bool
	IsAdequate          bool
	Message             string
}

// Design calculates required reinforcement for a doubly reinforced beam
func (b *DoublyReinforced) Design(mu float64) (*DoublyDesignResult, error) {
	b.Mu = mu

	if b.Width <= 0 || b.EffectiveDepth <= 0 {
		return nil, fmt.Errorf("invalid beam dimensions: width=%.2f, d=%.2f", b.Width, b.EffectiveDepth)
	}
	if b.Fc <= 0 || b.Fy <= 0 {
		return nil, fmt.Errorf("invalid material properties: f'c=%.2f, fy=%.2f", b.Fc, b.Fy)
	}
	if b.CoverComp <= 0 {
		return nil, fmt.Errorf("invalid compression cover: d'=%.2f", b.CoverComp)
	}

	result := &DoublyDesignResult{}
	beta1 := nscp.Beta1(b.Fc)

	// Calculate reinforcement ratio limits
	result.RhoMin = nscp.RhoMin(b.Fc, b.Fy)
	result.RhoMax = nscp.RhoMax(b.Fc, b.Fy)
	result.RhoBalanced = nscp.RhoBalanced(b.Fc, b.Fy)

	result.AsMin = result.RhoMin * b.Width * b.EffectiveDepth
	result.AsMax = result.RhoMax * b.Width * b.EffectiveDepth

	// Calculate maximum moment for singly reinforced (tension-controlled)
	// Using ρmax which corresponds to εt = 0.005
	result.AMax = result.RhoMax * b.Fy * b.Width * b.EffectiveDepth / (0.85 * b.Fc * b.Width)
	result.CMax = result.AMax / beta1

	phi := nscp.PhiFlexure
	Mu1Max := phi * 0.85 * b.Fc * b.Width * result.AMax * (b.EffectiveDepth - result.AMax/2) / 1e6

	// Convert Mu from kN-m to N-mm
	muNmm := mu * 1e6

	// Check if singly reinforced is adequate
	if mu <= Mu1Max {
		// Singly reinforced is adequate
		result.RequiresCompSteel = false
		result.Mu1 = mu
		result.Mu2 = 0

		// Calculate required steel using singly reinforced approach
		Rn := muNmm / (phi * b.Width * math.Pow(b.EffectiveDepth, 2))
		term := 2 * Rn / (0.85 * b.Fc)
		rhoRequired := (0.85 * b.Fc / b.Fy) * (1 - math.Sqrt(1-term))

		if rhoRequired < result.RhoMin {
			rhoRequired = result.RhoMin
		}

		result.As1 = rhoRequired * b.Width * b.EffectiveDepth
		result.As2 = 0
		result.AsTotal = result.As1
		result.AscRequired = 0

		// Calculate section properties
		a := result.AsTotal * b.Fy / (0.85 * b.Fc * b.Width)
		c := a / beta1
		result.EpsilonT = nscp.EpsilonCU * (b.EffectiveDepth - c) / c
		result.Phi = nscp.Phi(result.EpsilonT, b.Fy)
		result.IsTensionControlled = result.EpsilonT >= 0.005

		result.PhiMn = result.Phi * result.AsTotal * b.Fy * (b.EffectiveDepth - a/2) / 1e6
		result.IsAdequate = true
		result.Message = "Singly reinforced design is adequate"

		return result, nil
	}

	// Doubly reinforced design required
	result.RequiresCompSteel = true
	result.Mu1 = Mu1Max
	result.Mu2 = mu - Mu1Max

	// As1 = steel to resist Mu1 (at maximum for tension-controlled)
	result.As1 = result.AsMax

	// Check if compression steel yields
	// εsc = εcu * (c - d') / c
	result.EpsilonSc = nscp.EpsilonCU * (result.CMax - b.CoverComp) / result.CMax
	epsilonY := b.Fy / nscp.Es

	if result.EpsilonSc >= epsilonY {
		// Compression steel yields
		result.CompYielded = true
		result.FscStress = b.Fy
	} else {
		// Compression steel does not yield
		result.CompYielded = false
		result.FscStress = result.EpsilonSc * nscp.Es
	}

	// Mu2 is resisted by the steel couple
	// Mu2 = φ * As2 * fy * (d - d')
	// As2 = Mu2 / (φ * fy * (d - d'))
	leverArm := b.EffectiveDepth - b.CoverComp
	result.As2 = (result.Mu2 * 1e6) / (phi * b.Fy * leverArm)

	result.AsTotal = result.As1 + result.As2

	// Required compression steel
	// Force equilibrium: As2 * fy = Asc * fsc
	result.AscRequired = result.As2 * b.Fy / result.FscStress

	// For doubly reinforced at ρmax, the section is at the tension-controlled limit
	// εt = 0.005, so φ = 0.90
	result.EpsilonT = 0.005 // At the tension-controlled limit by design
	result.Phi = nscp.PhiFlexure
	result.IsTensionControlled = true

	// Calculate capacity
	// Mn = Mn1 + Mn2 where:
	// Mn1 = As1 * fy * (d - a/2) - moment from concrete couple
	// Mn2 = As2 * fy * (d - d') - moment from steel couple
	Mn1 := result.As1 * b.Fy * (b.EffectiveDepth - result.AMax/2)
	Mn2 := result.As2 * b.Fy * leverArm
	result.PhiMn = result.Phi * (Mn1 + Mn2) / 1e6

	result.IsAdequate = result.PhiMn >= mu*0.999 // Small tolerance for floating point

	if result.IsAdequate {
		if result.CompYielded {
			result.Message = "Doubly reinforced design OK - Compression steel yields"
		} else {
			result.Message = fmt.Sprintf("Doubly reinforced design OK - Compression steel does not yield (f'sc = %.1f MPa)", result.FscStress)
		}
	} else {
		result.Message = "Design inadequate - Consider increasing section size"
	}

	return result, nil
}

// DoublyAnalysisResult holds the results of doubly reinforced beam analysis
type DoublyAnalysisResult struct {
	// Section properties
	A     float64 // Depth of compression block (mm)
	C     float64 // Neutral axis depth (mm)
	Beta1 float64 // Stress block factor

	// Strains
	EpsilonT  float64 // Tensile strain
	EpsilonSc float64 // Compression steel strain

	// Stresses
	FsStress  float64 // Tension steel stress (MPa)
	FscStress float64 // Compression steel stress (MPa)

	// Steel yielding status
	TensionYielded bool
	CompYielded    bool

	// Reinforcement ratios
	Rho         float64
	RhoComp     float64
	RhoMin      float64
	RhoMax      float64
	RhoBalanced float64

	// Forces (kN)
	Cc float64 // Concrete compression force
	Cs float64 // Compression steel force
	T  float64 // Tension steel force

	// Capacity
	Phi   float64 // Strength reduction factor
	Mn    float64 // Nominal moment capacity (kN-m)
	PhiMn float64 // Design moment capacity (kN-m)

	// Status
	IsTensionControlled bool
	MeetsMinReinf       bool
	Message             string
}

// Analyze calculates moment capacity for a doubly reinforced beam
func (b *DoublyReinforced) Analyze(as, asc float64) (*DoublyAnalysisResult, error) {
	b.As = as
	b.Asc = asc

	if b.Width <= 0 || b.EffectiveDepth <= 0 {
		return nil, fmt.Errorf("invalid beam dimensions: width=%.2f, d=%.2f", b.Width, b.EffectiveDepth)
	}
	if b.Fc <= 0 || b.Fy <= 0 {
		return nil, fmt.Errorf("invalid material properties: f'c=%.2f, fy=%.2f", b.Fc, b.Fy)
	}
	if as <= 0 {
		return nil, fmt.Errorf("invalid tension reinforcement: As=%.2f", as)
	}
	if asc < 0 {
		return nil, fmt.Errorf("invalid compression reinforcement: A'sc=%.2f", asc)
	}

	result := &DoublyAnalysisResult{}
	result.Beta1 = nscp.Beta1(b.Fc)

	// Calculate reinforcement ratio limits
	result.RhoMin = nscp.RhoMin(b.Fc, b.Fy)
	result.RhoMax = nscp.RhoMax(b.Fc, b.Fy)
	result.RhoBalanced = nscp.RhoBalanced(b.Fc, b.Fy)

	// Actual reinforcement ratios
	result.Rho = as / (b.Width * b.EffectiveDepth)
	result.RhoComp = asc / (b.Width * b.EffectiveDepth)
	result.MeetsMinReinf = result.Rho >= result.RhoMin

	epsilonY := b.Fy / nscp.Es

	// Iterative solution to find neutral axis depth c
	// Force equilibrium: T = Cc + Cs
	// As*fs = 0.85*f'c*b*a + Asc*(fsc - 0.85*f'c)
	// Note: We subtract 0.85*f'c from compression steel stress to account for
	// displaced concrete (only if compression steel is within the stress block)

	// Initial guess: assume both steels yield
	// As*fy = 0.85*f'c*b*β1*c + Asc*(fy - 0.85*f'c)
	c := (as*b.Fy - asc*(b.Fy-0.85*b.Fc)) / (0.85 * b.Fc * b.Width * result.Beta1)

	// Iterate to find correct c
	for i := 0; i < 50; i++ {
		// Calculate strains
		epsilonT := nscp.EpsilonCU * (b.EffectiveDepth - c) / c
		epsilonSc := nscp.EpsilonCU * (c - b.CoverComp) / c

		// Calculate stresses
		fs := math.Min(epsilonT*nscp.Es, b.Fy)
		if epsilonT < 0 {
			fs = 0 // Should not happen for properly reinforced beam
		}

		var fsc float64
		if epsilonSc > 0 {
			fsc = math.Min(epsilonSc*nscp.Es, b.Fy)
		} else {
			fsc = math.Max(epsilonSc*nscp.Es, -b.Fy) // Compression steel in tension
		}

		// Recalculate c based on force equilibrium
		a := result.Beta1 * c
		
		// Account for displaced concrete if compression steel is within stress block
		var fscNet float64
		if a >= b.CoverComp {
			fscNet = fsc - 0.85*b.Fc // Subtract displaced concrete
		} else {
			fscNet = fsc
		}

		cNew := (as*fs - asc*fscNet) / (0.85 * b.Fc * b.Width * result.Beta1)

		if math.Abs(cNew-c) < 0.01 {
			c = cNew
			break
		}
		c = (c + cNew) / 2 // Damped iteration
	}

	result.C = c
	result.A = result.Beta1 * c

	// Final strains and stresses
	result.EpsilonT = nscp.EpsilonCU * (b.EffectiveDepth - c) / c
	result.EpsilonSc = nscp.EpsilonCU * (c - b.CoverComp) / c

	result.FsStress = math.Min(result.EpsilonT*nscp.Es, b.Fy)
	result.TensionYielded = result.EpsilonT >= epsilonY

	if result.EpsilonSc > 0 {
		result.FscStress = math.Min(result.EpsilonSc*nscp.Es, b.Fy)
		result.CompYielded = result.EpsilonSc >= epsilonY
	} else {
		result.FscStress = 0
		result.CompYielded = false
	}

	// Calculate forces (in kN)
	result.Cc = 0.85 * b.Fc * b.Width * result.A / 1000
	
	// Net compression steel force (accounting for displaced concrete)
	var fscNet float64
	if result.A >= b.CoverComp {
		fscNet = result.FscStress - 0.85*b.Fc
	} else {
		fscNet = result.FscStress
	}
	result.Cs = asc * fscNet / 1000
	result.T = as * result.FsStress / 1000

	// Strength reduction factor
	result.Phi = nscp.Phi(result.EpsilonT, b.Fy)
	result.IsTensionControlled = result.EpsilonT >= 0.005

	// Calculate moment capacity
	// Mn = Cc*(d - a/2) + Cs*(d - d')
	Mn := result.Cc*(b.EffectiveDepth-result.A/2) + result.Cs*(b.EffectiveDepth-b.CoverComp)
	result.Mn = Mn / 1000   // Convert to kN-m
	result.PhiMn = result.Phi * result.Mn

	// Build status message
	if result.IsTensionControlled {
		result.Message = "Section is tension-controlled (εt ≥ 0.005)"
	} else if result.EpsilonT >= epsilonY {
		result.Message = "Section is in transition zone"
	} else {
		result.Message = "Section is compression-controlled (εt < εy)"
	}

	if !result.MeetsMinReinf {
		result.Message += " | WARNING: Below minimum reinforcement"
	}

	return result, nil
}

