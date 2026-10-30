// Package geometry builds the printable lamp parts (lampshade and base)
// as signed distance fields. Callers can dot-import this package together
// with design to use both unqualified.
package geometry

import (
	"fmt"
	"math"

	"github.com/deadsy/sdfx/sdf"
	"github.com/deadsy/sdfx/vec/v3"

	"lamp3d-cli/internal/design"
)

// deg2rad converts degrees to radians.
func deg2rad(deg float64) float64 { return deg * math.Pi / 180 }

// shadeSDF is a custom signed distance field (SDF) for the twisted wave
// lampshade, similar to the reference "wave lamp" design.
//
// The silhouette radius smoothly varies with height (the large wavy bulges),
// and a surface pattern runs around the shade depending on the design mode
// (spiral ribs, flutes, rings or lobes). The bottom is closed with a plate
// that has a central hole for the light of the LED unit; the top stays open.
type shadeSDF struct {
	p         *design.Params // design parameters (mode, height, waves, ribs, wall)
	maxRadius float64        // conservative outer bound for the bounding box
}

// NewShade validates the parameters and builds the lampshade SDF.
func NewShade(p *design.Params) (sdf.SDF3, error) {
	if p.RibCount < 0 {
		return nil, fmt.Errorf("rib count must not be negative, got %d", p.RibCount)
	}
	if p.WallThickness <= 0 {
		return nil, fmt.Errorf("wall thickness must be positive, got %v", p.WallThickness)
	}
	return &shadeSDF{
		p: p,
		maxRadius: p.ShadeRadius +
			math.Abs(p.WaveAmp1) + math.Abs(p.WaveAmp2) +
			p.RibAmplitude + p.DuneAmplitude + p.WallThickness,
	}, nil
}

// surfaceRadius returns the outer surface radius at normalized height u and
// angle theta, including the mode-specific surface pattern.
func (s *shadeSDF) surfaceRadius(u, theta float64) float64 {
	p := s.p
	base := design.SilhouetteRadius(p, u)
	accTwist := deg2rad(p.TwistDegrees) * u * p.ShadeHeight // accumulated twist in radians

	switch p.Mode {
	case design.ModeSpiral:
		// A few bold helical ribs winding around the shade.
		rib := p.RibAmplitude * math.Cos(float64(p.SpiralArms)*(theta+deg2rad(p.TwistDegrees)*u))
		return base + rib
	case design.ModeRipple:
		// Horizontal rings stacked along the height.
		ring := p.RibAmplitude * math.Cos(2*math.Pi*float64(p.RingCount)*u)
		return base + ring
	case design.ModeDune:
		// Fine spiral ribs plus large vertical lobes drifting sideways.
		rib := p.RibAmplitude * math.Cos(float64(p.RibCount)*theta+accTwist)
		lobe := p.DuneAmplitude * math.Cos(float64(p.DuneLobes)*theta+deg2rad(p.WavePhase2)+accTwist*0.5)
		return base + rib + lobe
	default: // design.ModeWave and design.ModeFluted share the ribbed surface.
		rib := p.RibAmplitude * math.Cos(float64(p.RibCount)*theta+accTwist)
		return base + rib
	}
}

// Evaluate implements sdf.SDF3. It returns negative values inside the
// material and positive values outside. The shade is a thin shell with a
// closed bottom plate (with a central hole) and an open top rim. The
// radial-distance approximation keeps the evaluation cheap; small
// deviations on steep slopes are acceptable for a decorative lampshade.
func (s *shadeSDF) Evaluate(p v3.Vec) float64 {
	r := math.Hypot(p.X, p.Y)
	theta := math.Atan2(p.Y, p.X)
	z := p.Z
	wall := s.p.WallThickness

	// --- Shell wall: signed radial distance to the wall band
	// (the wall spans [outer-wall, outer] around the patterned surface). ---
	u := math.Min(math.Max(z/s.p.ShadeHeight, 0), 1)
	outer := s.surfaceRadius(u, theta)
	dWall := math.Abs(r-(outer-wall/2)) - wall/2

	// --- Bottom plate: flat disc from the wall in to a central hole. ---
	innerR := s.surfaceRadius(0, theta) - wall
	holeR := s.p.BottomHoleDiameter / 2
	plateT := wall
	dPlateR := math.Max(holeR-r, r-innerR)     // radial distance to the plate band
	dPlateZ := math.Abs(z-plateT/2) - plateT/2 // axial distance to the plate slab
	dPlate := math.Max(dPlateR, dPlateZ)       // plate = band intersected with slab

	d := math.Min(dWall, dPlate)

	// --- Ends: open top rim, closed bottom plate. ---
	if z > s.p.ShadeHeight {
		// Above the top: nearest material is the top rim annulus.
		return math.Hypot(z-s.p.ShadeHeight, math.Max(dWall, 0))
	}
	if z < 0 {
		// Below the bottom: nearest material is the plate disc.
		return math.Hypot(-z, math.Max(dPlateR, 0))
	}
	return d
}

// BoundingBox implements sdf.SDF3. The shade spans z in [0, ShadeHeight].
func (s *shadeSDF) BoundingBox() sdf.Box3 {
	return sdf.NewBox3(
		v3.Vec{X: 0, Y: 0, Z: s.p.ShadeHeight / 2},
		v3.Vec{X: 2 * s.maxRadius, Y: 2 * s.maxRadius, Z: s.p.ShadeHeight},
	)
}
