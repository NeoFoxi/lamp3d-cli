package main

import (
	"fmt"
	"math"

	"github.com/deadsy/sdfx/sdf"
	"github.com/deadsy/sdfx/vec/v3"
)

// SilhouetteRadius returns the average outer radius of the shade at
// normalized height u (0 = bottom, 1 = top). Two overlapping sine waves
// create the organic wavy outline seen in the reference photo.
func SilhouetteRadius(p *Params, u float64) float64 {
	return p.ShadeRadius +
		p.WaveAmp1*math.Sin(2*math.Pi*u+deg2rad(p.WavePhase1)) +
		p.WaveAmp2*math.Sin(2*math.Pi*2.3*u+deg2rad(p.WavePhase2))
}

// deg2rad converts degrees to radians.
func deg2rad(deg float64) float64 { return deg * math.Pi / 180 }

// shadeSDF is a custom signed distance field (SDF) for the twisted wave
// lampshade, similar to the reference "wave lamp" design.
//
// The silhouette radius smoothly varies with height (the large wavy bulges),
// and a rib pattern runs around the circumference, spiraling upward with the
// twist. The shade is a thin open-ended shell that is printed in vase mode.
type shadeSDF struct {
	p         *Params // design parameters (height, waves, ribs, wall)
	twistRad  float64 // total twist in radians
	maxRadius float64 // conservative outer bound for the bounding box
}

// NewShade validates the parameters and builds the lampshade SDF.
func NewShade(p *Params) (sdf.SDF3, error) {
	if p.RibCount < 3 {
		return nil, fmt.Errorf("rib count must be at least 3, got %d", p.RibCount)
	}
	if p.WallThickness <= 0 {
		return nil, fmt.Errorf("wall thickness must be positive, got %v", p.WallThickness)
	}
	return &shadeSDF{
		p:        p,
		twistRad: deg2rad(p.TwistDegrees),
		maxRadius: p.ShadeRadius +
			math.Abs(p.WaveAmp1) + math.Abs(p.WaveAmp2) +
			p.RibAmplitude + p.WallThickness,
	}, nil
}

// outerRadius returns the outer surface radius at height z and angle theta,
// including the spiral rib pattern.
func (s *shadeSDF) outerRadius(z, theta float64) float64 {
	u := math.Min(math.Max(z/s.p.ShadeHeight, 0), 1)
	rib := s.p.RibAmplitude * math.Cos(float64(s.p.RibCount)*theta+s.twistRad*z)
	return SilhouetteRadius(s.p, u) + rib
}

// ShadeBottomMidRadius returns the mid-wall radius of the shade at its
// bottom rim. The base uses this to size the alignment groove for the shade.
func ShadeBottomMidRadius(p *Params) float64 {
	return SilhouetteRadius(p, 0) - p.WallThickness/2
}

// Evaluate implements sdf.SDF3. It returns negative values inside the shell
// wall and positive values outside. The radial-distance approximation keeps
// the evaluation cheap; small deviations on steep slopes are acceptable for
// a decorative lampshade.
func (s *shadeSDF) Evaluate(p v3.Vec) float64 {
	r := math.Hypot(p.X, p.Y)
	theta := math.Atan2(p.Y, p.X)

	// Signed radial distance to the shell wall (the wall spans
	// [outer-wall, outer] around the ribbed surface).
	zClamped := math.Min(math.Max(p.Z, 0), s.p.ShadeHeight)
	outer := s.outerRadius(zClamped, theta)
	d := math.Abs(r-(outer-s.p.WallThickness/2)) - s.p.WallThickness/2

	// Distance to the flat open ends of the shade.
	dz := math.Min(p.Z, s.p.ShadeHeight-p.Z) // positive inside the axial span
	if dz < 0 {
		// Outside the axial span: distance to the rim annulus at the
		// nearest end plane.
		return math.Hypot(-dz, math.Max(d, 0))
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
