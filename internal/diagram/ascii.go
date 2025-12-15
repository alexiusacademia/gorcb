package diagram

import (
	"fmt"
	"strings"
)

// Point represents a 2D coordinate for section vertices
type Point struct {
	X float64
	Y float64
}

// SectionDiagramData holds data for drawing a beam section diagram
type SectionDiagramData struct {
	// Beam dimensions
	Width  float64 // mm
	Height float64 // mm

	// Custom section vertices (if provided, draws actual shape)
	Vertices []Point // Counter-clockwise from bottom-left

	// Analysis results
	NeutralAxisDepth float64 // c - from top (mm)
	StressBlockDepth float64 // a - from top (mm)

	// Reinforcement
	TensionSteelY     float64 // Distance from bottom (mm)
	TensionSteelArea  float64 // mm²
	CompSteelY        float64 // Distance from bottom (mm), 0 if none
	CompSteelArea     float64 // mm², 0 if none

	// Strains
	EpsilonCU float64 // Concrete ultimate strain (typically 0.003)
	EpsilonT  float64 // Tension steel strain
	EpsilonSC float64 // Compression steel strain (if applicable)
	EpsilonY  float64 // Yield strain

	// Stresses (MPa)
	Fc        float64 // Concrete stress (0.85 f'c)
	FsTension float64 // Tension steel stress
	FsComp    float64 // Compression steel stress

	// Status
	TensionYields bool
	CompYields    bool
	IsDoubly      bool
}

// DrawASCIISectionDiagram creates an ASCII representation of beam section with stress block
func DrawASCIISectionDiagram(data SectionDiagramData) string {
	var sb strings.Builder

	// Scale factors for ASCII drawing
	widthChars := 30
	heightChars := 20

	// Calculate proportions
	naRatio := data.NeutralAxisDepth / data.Height
	aRatio := data.StressBlockDepth / data.Height
	tensionY := data.TensionSteelY / data.Height
	compY := data.CompSteelY / data.Height

	naLine := int(naRatio * float64(heightChars))
	aLine := int(aRatio * float64(heightChars))
	tensionLine := heightChars - int(tensionY*float64(heightChars))
	compLine := heightChars - int(compY*float64(heightChars))

	sb.WriteString("\n")
	sb.WriteString("  BEAM SECTION                    STRAIN              STRESS\n")
	sb.WriteString("  ────────────                    ──────              ──────\n")

	for i := 0; i <= heightChars; i++ {
		// Section column
		if i == 0 {
			sb.WriteString(fmt.Sprintf("  ┌%s┐", strings.Repeat("─", widthChars)))
		} else if i == heightChars {
			sb.WriteString(fmt.Sprintf("  └%s┘", strings.Repeat("─", widthChars)))
		} else {
			// Fill with stress block shading
			var fill string
			if i <= aLine {
				// Stress block region (compression)
				fill = strings.Repeat("░", widthChars)
			} else {
				fill = strings.Repeat(" ", widthChars)
			}

			// Add compression steel marker
			if data.IsDoubly && i == compLine {
				mid := widthChars / 2
				fill = fill[:mid-2] + "●──●" + fill[mid+2:]
			}

			// Add tension steel marker
			if i == tensionLine {
				mid := widthChars / 2
				if widthChars >= 10 {
					fill = fill[:mid-3] + "●────●" + fill[mid+3:]
				}
			}

			// Neutral axis marker
			if i == naLine {
				sb.WriteString(fmt.Sprintf("  │%s│", fill))
				sb.WriteString(" ◄─ N.A.")
			} else {
				sb.WriteString(fmt.Sprintf("  │%s│", fill))
			}
		}

		// Strain diagram column
		sb.WriteString("    ")
		if i == 0 {
			sb.WriteString(fmt.Sprintf("  ├── εcu = %.4f", data.EpsilonCU))
		} else if i == naLine {
			sb.WriteString("  ├── ε = 0")
		} else if i == tensionLine {
			yieldMark := ""
			if data.TensionYields {
				yieldMark = " (yields)"
			}
			sb.WriteString(fmt.Sprintf("  ├── εt = %.4f%s", data.EpsilonT, yieldMark))
		} else if data.IsDoubly && i == compLine {
			yieldMark := ""
			if data.CompYields {
				yieldMark = " (yields)"
			}
			sb.WriteString(fmt.Sprintf("  ├── ε'sc = %.4f%s", data.EpsilonSC, yieldMark))
		} else if i > 0 && i < heightChars {
			if i < naLine {
				sb.WriteString("  │")
			} else {
				sb.WriteString("  │")
			}
		}

		// Stress diagram column
		if i == 0 {
			sb.WriteString(fmt.Sprintf("      ┌── 0.85f'c = %.1f MPa", data.Fc))
		} else if i == aLine && aLine > 0 {
			sb.WriteString("      └── (stress block)")
		} else if i == tensionLine {
			sb.WriteString(fmt.Sprintf("      ── fs = %.1f MPa", data.FsTension))
		} else if data.IsDoubly && i == compLine {
			sb.WriteString(fmt.Sprintf("      ── f'sc = %.1f MPa", data.FsComp))
		}

		sb.WriteString("\n")
	}

	// Legend
	sb.WriteString("\n")
	sb.WriteString("  Legend:\n")
	sb.WriteString("  ░░░ = Compression zone (stress block)\n")
	sb.WriteString("  ●●● = Reinforcement\n")
	sb.WriteString(fmt.Sprintf("  N.A. = Neutral Axis at c = %.1f mm from top\n", data.NeutralAxisDepth))
	sb.WriteString(fmt.Sprintf("  Stress block depth a = %.1f mm\n", data.StressBlockDepth))

	return sb.String()
}

// DrawStrainDiagram creates an ASCII strain distribution diagram
func DrawStrainDiagram(data SectionDiagramData) string {
	var sb strings.Builder

	height := 15
	width := 40

	// Scale strains to fit
	maxStrain := max(data.EpsilonCU, data.EpsilonT)
	scale := float64(width-10) / maxStrain

	sb.WriteString("\n")
	sb.WriteString("  STRAIN DISTRIBUTION DIAGRAM\n")
	sb.WriteString("  ───────────────────────────\n\n")

	// Calculate positions
	naRatio := data.NeutralAxisDepth / data.Height
	naLine := int(naRatio * float64(height))
	tensionLine := height - int((data.TensionSteelY/data.Height)*float64(height))

	for i := 0; i <= height; i++ {
		// Calculate strain at this depth
		depthRatio := float64(i) / float64(height)
		depth := depthRatio * data.Height
		
		var strain float64
		if depth <= data.NeutralAxisDepth {
			// Compression zone - positive strain
			strain = data.EpsilonCU * (data.NeutralAxisDepth - depth) / data.NeutralAxisDepth
		} else {
			// Tension zone - negative strain (show as positive for diagram)
			strain = data.EpsilonCU * (depth - data.NeutralAxisDepth) / data.NeutralAxisDepth
		}

		barLen := int(strain * scale)
		if barLen < 0 {
			barLen = 0
		}

		// Draw the bar
		if i == 0 {
			sb.WriteString(fmt.Sprintf("  Top    │%s▶ εcu=%.4f\n", strings.Repeat("█", barLen), data.EpsilonCU))
		} else if i == naLine {
			sb.WriteString(fmt.Sprintf("  N.A.   ├%s (ε=0)\n", strings.Repeat("─", 5)))
		} else if i == tensionLine {
			mark := ""
			if data.TensionYields {
				mark = " ✓yields"
			}
			sb.WriteString(fmt.Sprintf("  Steel  │%s▶ εt=%.4f%s\n", strings.Repeat("█", barLen), data.EpsilonT, mark))
		} else if i == height {
			sb.WriteString(fmt.Sprintf("  Bottom │%s\n", strings.Repeat("█", barLen)))
		} else {
			sb.WriteString(fmt.Sprintf("         │%s\n", strings.Repeat("█", barLen)))
		}
	}

	// Yield strain reference
	yieldBar := int(data.EpsilonY * scale)
	sb.WriteString(fmt.Sprintf("\n  εy = %.4f %s (yield strain)\n", data.EpsilonY, strings.Repeat("─", yieldBar)+"┤"))

	return sb.String()
}

// DrawStressBlock creates a simple stress block diagram
func DrawStressBlock(data SectionDiagramData) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("  EQUIVALENT RECTANGULAR STRESS BLOCK\n")
	sb.WriteString("  ────────────────────────────────────\n\n")

	// Simplified view
	sb.WriteString("       ┌───────────────┐\n")
	sb.WriteString("       │               │\n")
	sb.WriteString(fmt.Sprintf("       │  0.85f'c      │ ← a = %.1f mm\n", data.StressBlockDepth))
	sb.WriteString(fmt.Sprintf("       │  = %.1f MPa   │\n", data.Fc))
	sb.WriteString("       │               │\n")
	sb.WriteString("       └───────────────┘ ─── Cc = 0.85·f'c·b·a\n")
	sb.WriteString("       ─ ─ ─ ─ ─ ─ ─ ─ ─ ← N.A. (c = %.1f mm)\n")
	sb.WriteString("                         │\n")
	sb.WriteString("                         │  (d - a/2)\n")
	sb.WriteString("                         │\n")
	sb.WriteString("       ●═══════════════● ← Tension Steel\n")
	sb.WriteString(fmt.Sprintf("         As = %.1f mm²\n", data.TensionSteelArea))
	sb.WriteString(fmt.Sprintf("         T = As·fs = %.1f kN\n", data.TensionSteelArea*data.FsTension/1000))

	return sb.String()
}

// DrawSummaryBox creates a summary box for results
func DrawSummaryBox(title string, lines []string) string {
	var sb strings.Builder

	maxLen := len(title)
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	maxLen += 4

	border := strings.Repeat("═", maxLen)
	sb.WriteString(fmt.Sprintf("  ╔%s╗\n", border))
	sb.WriteString(fmt.Sprintf("  ║  %-*s  ║\n", maxLen-2, title))
	sb.WriteString(fmt.Sprintf("  ╠%s╣\n", border))
	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("  ║  %-*s  ║\n", maxLen-2, line))
	}
	sb.WriteString(fmt.Sprintf("  ╚%s╝\n", border))

	return sb.String()
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

