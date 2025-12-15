package diagram

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// ExportSectionDiagram exports a beam section diagram to an image file
func ExportSectionDiagram(data SectionDiagramData, filename string) error {
	p := plot.New()
	p.Title.Text = "Beam Section Analysis"
	p.X.Label.Text = "Width (mm)"
	p.Y.Label.Text = "Height (mm)"

	var minX, maxX float64

	// Check if we have custom vertices (non-rectangular section)
	if len(data.Vertices) >= 3 {
		// Draw custom section outline
		beamOutline := make(plotter.XYs, len(data.Vertices)+1)
		minX, maxX = data.Vertices[0].X, data.Vertices[0].X
		
		for i, v := range data.Vertices {
			beamOutline[i] = plotter.XY{X: v.X, Y: v.Y}
			if v.X < minX {
				minX = v.X
			}
			if v.X > maxX {
				maxX = v.X
			}
		}
		beamOutline[len(data.Vertices)] = plotter.XY{X: data.Vertices[0].X, Y: data.Vertices[0].Y}

		beamLine, err := plotter.NewLine(beamOutline)
		if err != nil {
			return err
		}
		beamLine.LineStyle.Width = vg.Points(2)
		beamLine.LineStyle.Color = color.Black
		p.Add(beamLine)

		// Draw stress block by clipping section at stress block depth
		stressBlockPts := clipSectionAtDepth(data.Vertices, data.Height, data.StressBlockDepth)
		if len(stressBlockPts) >= 3 {
			stressBlock, err := plotter.NewPolygon(stressBlockPts)
			if err == nil {
				stressBlock.Color = color.RGBA{R: 100, G: 149, B: 237, A: 150}
				stressBlock.LineStyle.Color = color.RGBA{R: 0, G: 0, B: 139, A: 255}
				p.Add(stressBlock)
			}
		}
	} else {
		// Draw rectangular beam outline
		minX, maxX = 0, data.Width
		beamOutline := plotter.XYs{
			{X: 0, Y: 0},
			{X: data.Width, Y: 0},
			{X: data.Width, Y: data.Height},
			{X: 0, Y: data.Height},
			{X: 0, Y: 0},
		}
		beamLine, err := plotter.NewLine(beamOutline)
		if err != nil {
			return err
		}
		beamLine.LineStyle.Width = vg.Points(2)
		beamLine.LineStyle.Color = color.Black
		p.Add(beamLine)

		// Draw rectangular stress block
		stressBlockPts := plotter.XYs{
			{X: 0, Y: data.Height},
			{X: data.Width, Y: data.Height},
			{X: data.Width, Y: data.Height - data.StressBlockDepth},
			{X: 0, Y: data.Height - data.StressBlockDepth},
		}
		stressBlock, err := plotter.NewPolygon(stressBlockPts)
		if err != nil {
			return err
		}
		stressBlock.Color = color.RGBA{R: 100, G: 149, B: 237, A: 150}
		stressBlock.LineStyle.Color = color.RGBA{R: 0, G: 0, B: 139, A: 255}
		p.Add(stressBlock)
	}

	sectionWidth := maxX - minX

	// Draw neutral axis line
	naY := data.Height - data.NeutralAxisDepth
	naLine, err := plotter.NewLine(plotter.XYs{
		{X: minX - 20, Y: naY},
		{X: maxX + 20, Y: naY},
	})
	if err != nil {
		return err
	}
	naLine.LineStyle.Width = vg.Points(1.5)
	naLine.LineStyle.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	naLine.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(3)}
	p.Add(naLine)

	// Find web width at tension steel level for positioning
	webMinX, webMaxX := findWidthAtY(data.Vertices, data.TensionSteelY, minX, maxX)
	webCenter := (webMinX + webMaxX) / 2
	webWidth := webMaxX - webMinX
	if webWidth == 0 {
		webWidth = sectionWidth
		webCenter = (minX + maxX) / 2
	}

	// Draw tension steel
	tensionY := data.TensionSteelY
	tensionSteel, err := plotter.NewScatter(plotter.XYs{
		{X: webCenter - webWidth*0.2, Y: tensionY},
		{X: webCenter, Y: tensionY},
		{X: webCenter + webWidth*0.2, Y: tensionY},
	})
	if err != nil {
		return err
	}
	tensionSteel.GlyphStyle.Color = color.RGBA{R: 139, G: 69, B: 19, A: 255}
	tensionSteel.GlyphStyle.Radius = vg.Points(6)
	tensionSteel.GlyphStyle.Shape = draw.CircleGlyph{}
	p.Add(tensionSteel)

	// Draw compression steel if present
	if data.IsDoubly && data.CompSteelArea > 0 {
		compY := data.Height - data.CompSteelY
		compSteel, err := plotter.NewScatter(plotter.XYs{
			{X: webCenter - webWidth*0.15, Y: compY},
			{X: webCenter + webWidth*0.15, Y: compY},
		})
		if err != nil {
			return err
		}
		compSteel.GlyphStyle.Color = color.RGBA{R: 139, G: 69, B: 19, A: 255}
		compSteel.GlyphStyle.Radius = vg.Points(5)
		compSteel.GlyphStyle.Shape = draw.CircleGlyph{}
		p.Add(compSteel)
	}

	// Add annotations
	labels := []struct {
		x, y float64
		text string
	}{
		{maxX + 30, naY, "N.A."},
		{maxX + 30, data.Height - data.StressBlockDepth/2, fmt.Sprintf("a=%.1fmm", data.StressBlockDepth)},
		{webCenter, tensionY - 25, fmt.Sprintf("As=%.0fmm²", data.TensionSteelArea)},
	}

	for _, lbl := range labels {
		l, err := plotter.NewLabels(plotter.XYLabels{
			XYs:    []plotter.XY{{X: lbl.x, Y: lbl.y}},
			Labels: []string{lbl.text},
		})
		if err != nil {
			return err
		}
		p.Add(l)
	}

	// Determine file format from extension
	ext := filepath.Ext(filename)
	width := 8 * vg.Inch
	height := 6 * vg.Inch

	// Create directory if needed
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0755)
	}

	switch ext {
	case ".png":
		return p.Save(width, height, filename)
	case ".svg":
		return p.Save(width, height, filename)
	case ".pdf":
		return p.Save(width, height, filename)
	default:
		return p.Save(width, height, filename+".png")
	}
}

// clipSectionAtDepth clips the section polygon at a given depth from top
// Returns the vertices of the clipped (compression zone) polygon
func clipSectionAtDepth(vertices []Point, height, depth float64) plotter.XYs {
	if len(vertices) < 3 || depth <= 0 {
		return nil
	}

	clipY := height - depth
	var result plotter.XYs

	n := len(vertices)
	for i := 0; i < n; i++ {
		curr := vertices[i]
		next := vertices[(i+1)%n]

		currAbove := curr.Y >= clipY
		nextAbove := next.Y >= clipY

		if currAbove {
			result = append(result, plotter.XY{X: curr.X, Y: curr.Y})
		}

		// Check for intersection with clip line
		if currAbove != nextAbove {
			// Calculate intersection point
			t := (clipY - curr.Y) / (next.Y - curr.Y)
			intersectX := curr.X + t*(next.X-curr.X)
			result = append(result, plotter.XY{X: intersectX, Y: clipY})
		}
	}

	return result
}

// findWidthAtY finds the min and max X at a given Y level
func findWidthAtY(vertices []Point, y, defaultMin, defaultMax float64) (float64, float64) {
	if len(vertices) < 3 {
		return defaultMin, defaultMax
	}

	var intersections []float64
	n := len(vertices)

	for i := 0; i < n; i++ {
		curr := vertices[i]
		next := vertices[(i+1)%n]

		// Check if edge crosses this Y level
		if (curr.Y <= y && next.Y > y) || (next.Y <= y && curr.Y > y) {
			t := (y - curr.Y) / (next.Y - curr.Y)
			x := curr.X + t*(next.X-curr.X)
			intersections = append(intersections, x)
		}
	}

	if len(intersections) < 2 {
		return defaultMin, defaultMax
	}

	minX, maxX := intersections[0], intersections[0]
	for _, x := range intersections {
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
	}

	return minX, maxX
}

// ExportStrainDiagram exports a strain distribution diagram
func ExportStrainDiagram(data SectionDiagramData, filename string) error {
	p := plot.New()
	p.Title.Text = "Strain Distribution"
	p.X.Label.Text = "Strain"
	p.Y.Label.Text = "Depth from top (mm)"

	// Invert Y axis (depth increases downward)
	p.Y.Min = data.Height
	p.Y.Max = 0

	// Strain distribution line
	strainPts := plotter.XYs{
		{X: data.EpsilonCU, Y: 0},                           // Top - compression
		{X: 0, Y: data.NeutralAxisDepth},                     // Neutral axis
		{X: -data.EpsilonT, Y: data.Height - data.TensionSteelY}, // Tension steel level
	}
	strainLine, err := plotter.NewLine(strainPts)
	if err != nil {
		return err
	}
	strainLine.LineStyle.Width = vg.Points(2)
	strainLine.LineStyle.Color = color.RGBA{R: 0, G: 100, B: 0, A: 255}
	p.Add(strainLine)

	// Zero strain reference line
	zeroLine, err := plotter.NewLine(plotter.XYs{
		{X: 0, Y: 0},
		{X: 0, Y: data.Height},
	})
	if err != nil {
		return err
	}
	zeroLine.LineStyle.Width = vg.Points(1)
	zeroLine.LineStyle.Color = color.Gray{Y: 128}
	zeroLine.LineStyle.Dashes = []vg.Length{vg.Points(3), vg.Points(3)}
	p.Add(zeroLine)

	// Yield strain reference lines
	yieldLinePos, _ := plotter.NewLine(plotter.XYs{
		{X: data.EpsilonY, Y: 0},
		{X: data.EpsilonY, Y: data.Height},
	})
	yieldLinePos.LineStyle.Color = color.RGBA{R: 255, G: 165, B: 0, A: 255}
	yieldLinePos.LineStyle.Dashes = []vg.Length{vg.Points(2), vg.Points(2)}
	p.Add(yieldLinePos)

	yieldLineNeg, _ := plotter.NewLine(plotter.XYs{
		{X: -data.EpsilonY, Y: 0},
		{X: -data.EpsilonY, Y: data.Height},
	})
	yieldLineNeg.LineStyle.Color = color.RGBA{R: 255, G: 165, B: 0, A: 255}
	yieldLineNeg.LineStyle.Dashes = []vg.Length{vg.Points(2), vg.Points(2)}
	p.Add(yieldLineNeg)

	// Mark key points
	keyPoints, err := plotter.NewScatter(plotter.XYs{
		{X: data.EpsilonCU, Y: 0},
		{X: 0, Y: data.NeutralAxisDepth},
		{X: -data.EpsilonT, Y: data.Height - data.TensionSteelY},
	})
	if err != nil {
		return err
	}
	keyPoints.GlyphStyle.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	keyPoints.GlyphStyle.Radius = vg.Points(4)
	p.Add(keyPoints)

	width := 6 * vg.Inch
	height := 8 * vg.Inch

	// Create directory if needed
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0755)
	}

	return p.Save(width, height, filename)
}

// ExportCombinedDiagram creates a combined section and strain diagram
func ExportCombinedDiagram(data SectionDiagramData, filename string) error {
	// For now, just export the section diagram
	// A more sophisticated version could use subplots
	return ExportSectionDiagram(data, filename)
}

