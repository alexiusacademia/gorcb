package nscp

import "math"

// NSCP 2015 Material Constants

const (
	// Beta1 factors for equivalent rectangular stress block
	// Section 410.2.7.3
	Beta1Max = 0.85 // for f'c <= 28 MPa
	Beta1Min = 0.65 // minimum value

	// Strain limits
	EpsilonCU = 0.003 // Ultimate concrete strain (Section 410.2.2.1)
	EpsilonTY = 0.002 // Yield strain for Grade 60 steel (fy=415 MPa)

	// Strength reduction factors (Section 409.3.2)
	PhiFlexure       = 0.90 // Tension-controlled sections
	PhiShear         = 0.75 // Shear and torsion
	PhiCompression   = 0.65 // Compression-controlled (tied)
	PhiCompressionSp = 0.75 // Compression-controlled (spiral)

	// Modulus of elasticity for steel (Section 420.2.2)
	Es = 200000.0 // MPa
)

// Beta1 calculates the factor for equivalent rectangular stress block
// NSCP 2015 Section 410.2.7.3
func Beta1(fc float64) float64 {
	if fc <= 28 {
		return Beta1Max
	}
	// β1 = 0.85 - 0.05(f'c - 28)/7 for f'c > 28 MPa
	beta1 := Beta1Max - 0.05*(fc-28)/7
	return math.Max(beta1, Beta1Min)
}

// Phi calculates the strength reduction factor based on strain
// NSCP 2015 Section 409.3.2
func Phi(epsilonT float64, fy float64) float64 {
	epsilonTY := fy / Es

	if epsilonT >= epsilonTY+0.003 {
		// Tension-controlled
		return PhiFlexure
	} else if epsilonT <= epsilonTY {
		// Compression-controlled
		return PhiCompression
	}
	// Transition zone
	return PhiCompression + (PhiFlexure-PhiCompression)*(epsilonT-epsilonTY)/0.003
}

// RhoMin calculates minimum reinforcement ratio
// NSCP 2015 Section 409.6.1.2
func RhoMin(fc, fy float64) float64 {
	// ρmin = max(√f'c / 4fy, 1.4/fy)
	rho1 := math.Sqrt(fc) / (4 * fy)
	rho2 := 1.4 / fy
	return math.Max(rho1, rho2)
}

// RhoMax calculates maximum reinforcement ratio for tension-controlled section
// Based on strain compatibility for εt = 0.004 (tension-controlled limit)
func RhoMax(fc, fy float64) float64 {
	beta1 := Beta1(fc)
	// For tension-controlled: εt >= 0.005, use εt = 0.005
	// c/d = εcu / (εcu + εt) = 0.003 / (0.003 + 0.005) = 0.375
	// ρmax = 0.85 * β1 * (f'c/fy) * (0.003/(0.003+0.005))
	return 0.85 * beta1 * (fc / fy) * (EpsilonCU / (EpsilonCU + 0.005))
}

// RhoBalanced calculates balanced reinforcement ratio
func RhoBalanced(fc, fy float64) float64 {
	beta1 := Beta1(fc)
	epsilonTY := fy / Es
	// c/d at balanced = εcu / (εcu + εy)
	cb := EpsilonCU / (EpsilonCU + epsilonTY)
	return 0.85 * beta1 * (fc / fy) * cb
}

