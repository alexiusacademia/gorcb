package section

import (
	"math"
	"sort"
)

// CalculateProperties computes geometric properties of the section
func (s *Section) CalculateProperties() *SectionProperties {
	props := &SectionProperties{}

	if len(s.Vertices) < 3 {
		return props
	}

	// Find bounding box
	props.MinX, props.MaxX = s.Vertices[0].X, s.Vertices[0].X
	props.MinY, props.MaxY = s.Vertices[0].Y, s.Vertices[0].Y

	for _, v := range s.Vertices {
		props.MinX = math.Min(props.MinX, v.X)
		props.MaxX = math.Max(props.MaxX, v.X)
		props.MinY = math.Min(props.MinY, v.Y)
		props.MaxY = math.Max(props.MaxY, v.Y)
	}

	props.Width = props.MaxX - props.MinX
	props.Height = props.MaxY - props.MinY

	// Calculate area and centroid using the shoelace formula
	props.Area, props.CentroidX, props.CentroidY = s.calculateAreaAndCentroid()

	// Calculate reinforcement properties
	s.calculateReinforcementProperties(props)

	return props
}

// calculateAreaAndCentroid uses the shoelace formula
func (s *Section) calculateAreaAndCentroid() (area, cx, cy float64) {
	n := len(s.Vertices)
	if n < 3 {
		return 0, 0, 0
	}

	var signedArea float64
	var sumX, sumY float64

	for i := 0; i < n; i++ {
		j := (i + 1) % n
		cross := s.Vertices[i].X*s.Vertices[j].Y - s.Vertices[j].X*s.Vertices[i].Y
		signedArea += cross
		sumX += (s.Vertices[i].X + s.Vertices[j].X) * cross
		sumY += (s.Vertices[i].Y + s.Vertices[j].Y) * cross
	}

	signedArea /= 2
	area = math.Abs(signedArea)

	if area > 0 {
		cx = sumX / (6 * signedArea)
		cy = sumY / (6 * signedArea)
	}

	return area, cx, cy
}

// calculateReinforcementProperties calculates steel areas and effective depth
func (s *Section) calculateReinforcementProperties(props *SectionProperties) {
	if len(s.Reinforcement) == 0 {
		return
	}

	// Find the neutral axis estimate (mid-height for initial classification)
	midHeight := (props.MinY + props.MaxY) / 2

	var tensionArea, tensionMoment float64
	var compressionArea, compressionMoment float64

	for _, layer := range s.Reinforcement {
		if layer.Type == "compression" || (layer.Type == "" && layer.Y > midHeight) {
			// Compression steel (top of section)
			compressionArea += layer.Area
			compressionMoment += layer.Area * layer.Y
		} else {
			// Tension steel (bottom of section)
			tensionArea += layer.Area
			tensionMoment += layer.Area * layer.Y
		}
	}

	props.TotalTensionSteel = tensionArea
	props.TotalCompressionSteel = compressionArea

	// Effective depth to centroid of tension steel
	if tensionArea > 0 {
		tensionCentroid := tensionMoment / tensionArea
		props.EffectiveDepth = props.MaxY - tensionCentroid
	}

	// Cover to compression steel
	if compressionArea > 0 {
		compressionCentroid := compressionMoment / compressionArea
		props.CompressionCover = props.MaxY - compressionCentroid
	}

	// Override with user-specified effective depth if provided
	if s.EffectiveDepth > 0 {
		props.EffectiveDepth = s.EffectiveDepth
	}
}

// WidthAtDepth calculates the width of the section at a given depth from top
// Uses horizontal line intersection with the polygon
func (s *Section) WidthAtDepth(depthFromTop float64) float64 {
	props := s.CalculateProperties()
	y := props.MaxY - depthFromTop

	return s.widthAtY(y)
}

// widthAtY calculates the width at a specific Y coordinate
func (s *Section) widthAtY(y float64) float64 {
	intersections := s.findIntersectionsAtY(y)
	
	if len(intersections) < 2 {
		return 0
	}

	// Sort intersections by X coordinate
	sort.Float64s(intersections)

	// Total width is the sum of all segments
	var totalWidth float64
	for i := 0; i < len(intersections)-1; i += 2 {
		if i+1 < len(intersections) {
			totalWidth += intersections[i+1] - intersections[i]
		}
	}

	return totalWidth
}

// findIntersectionsAtY finds all X coordinates where a horizontal line at Y intersects the polygon
func (s *Section) findIntersectionsAtY(y float64) []float64 {
	var intersections []float64
	n := len(s.Vertices)

	for i := 0; i < n; i++ {
		j := (i + 1) % n
		v1, v2 := s.Vertices[i], s.Vertices[j]

		// Check if the edge crosses the Y level
		if (v1.Y <= y && v2.Y > y) || (v2.Y <= y && v1.Y > y) {
			// Calculate X at intersection
			t := (y - v1.Y) / (v2.Y - v1.Y)
			x := v1.X + t*(v2.X-v1.X)
			intersections = append(intersections, x)
		}
	}

	return intersections
}

// CompressionBlockArea calculates the area of the compression zone
// given the depth of the neutral axis from the top
func (s *Section) CompressionBlockArea(a float64) float64 {
	props := s.CalculateProperties()
	
	// Integrate width from top to depth a
	// Using numerical integration (trapezoidal rule)
	const numSteps = 100
	dy := a / float64(numSteps)
	
	var area float64
	for i := 0; i < numSteps; i++ {
		y1 := props.MaxY - float64(i)*dy
		y2 := props.MaxY - float64(i+1)*dy
		w1 := s.widthAtY(y1)
		w2 := s.widthAtY(y2)
		area += (w1 + w2) / 2 * dy
	}

	return area
}

// CompressionBlockCentroid calculates the centroid of the compression zone
// from the top of the section, given the depth of compression block a
func (s *Section) CompressionBlockCentroid(a float64) float64 {
	props := s.CalculateProperties()
	
	// Numerical integration to find centroid
	const numSteps = 100
	dy := a / float64(numSteps)
	
	var area, moment float64
	for i := 0; i < numSteps; i++ {
		y1 := props.MaxY - float64(i)*dy
		y2 := props.MaxY - float64(i+1)*dy
		yMid := (y1 + y2) / 2
		w1 := s.widthAtY(y1)
		w2 := s.widthAtY(y2)
		dA := (w1 + w2) / 2 * dy
		
		area += dA
		depthFromTop := props.MaxY - yMid
		moment += dA * depthFromTop
	}

	if area > 0 {
		return moment / area
	}
	return a / 2
}

