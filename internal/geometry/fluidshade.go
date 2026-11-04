package geometry

import (
	"fmt"
	"math"

	"github.com/deadsy/sdfx/sdf"
	"github.com/deadsy/sdfx/vec/v3"

	"lamp3d-cli/internal/fluid"
)

// fluidSDF is a custom SDF for the fluid-pattern lampshade: a simple smooth
// silhouette whose surface carries flowing patterns — layered twisted
// angular harmonics drifting in opposite directions, a breathing amplitude
// modulation along the height, and optional helically drifting rings.
//
// The silhouette radius smoothly varies with height (the large bulge), and
// the fluid surface pattern is layered on top of it. The drift fields are
// the total twist in degrees over the full shade height (sign = direction).
// The bottom is closed with a plate that has a central hole for the light
// of the LED unit; the top stays open.
type fluidSDF struct {
	f         *fluid.Params // fluid design parameters (embedded design dims + pattern knobs)
	maxRadius float64       // conservative outer bound for the bounding box
}

// NewFluidShade validates the parameters and builds the fluid-pattern lampshade SDF.
func NewFluidShade(f *fluid.Params) (sdf.SDF3, error) {
	if f.Harm1Freq < 0 {
		return nil, fmt.Errorf("harm1 frequency must not be negative, got %d", f.Harm1Freq)
	}
	if f.Harm2Freq < 0 {
		return nil, fmt.Errorf("harm2 frequency must not be negative, got %d", f.Harm2Freq)
	}
	if f.Harm1Amp < 0 {
		return nil, fmt.Errorf("harm1 amplitude must not be negative, got %v", f.Harm1Amp)
	}
	if f.Harm2Amp < 0 {
		return nil, fmt.Errorf("harm2 amplitude must not be negative, got %v", f.Harm2Amp)
	}
	if f.RingFreq < 0 {
		return nil, fmt.Errorf("ring frequency must not be negative, got %v", f.RingFreq)
	}
	if f.RingAmp < 0 {
		return nil, fmt.Errorf("ring amplitude must not be negative, got %v", f.RingAmp)
	}
	if f.BreathDepth < 0 || f.BreathDepth > 1 {
		return nil, fmt.Errorf("breath depth must be between 0 and 1, got %v", f.BreathDepth)
	}
	if f.WallThickness <= 0 {
		return nil, fmt.Errorf("wall thickness must be positive, got %v", f.WallThickness)
	}
	return &fluidSDF{
		f: f,
		maxRadius: f.ShadeRadius +
			math.Abs(f.BulgeAmp) +
			fluid.PatternAmpMax(f) +
			f.WallThickness,
	}, nil
}

// surfaceRadius returns the outer surface radius at normalized height u and
// angle theta, including the fluid surface pattern: two twisted angular
// harmonics drifting in opposite directions (modulated by the breathing
// factor) and optional helically drifting rings.
func (s *fluidSDF) surfaceRadius(u, theta float64) float64 {
	f := s.f
	base := fluid.SilhouetteRadius(f, u)

	// Breathing factor: compresses the harmonic amplitudes rhythmically
	// along the height; lies in [1-BreathDepth, 1].
	breath := 1.0
	if f.BreathFreq > 0 {
		breath = 1 - f.BreathDepth*(0.5+0.5*math.Cos(2*math.Pi*f.BreathFreq*u+deg2rad(f.BreathPhase)))
	}

	if f.Harm1Freq > 0 && f.Harm1Amp > 0 {
		base += f.Harm1Amp * breath * math.Cos(float64(f.Harm1Freq)*theta+deg2rad(f.Harm1Drift)*u+deg2rad(f.Harm1Phase))
	}
	if f.Harm2Freq > 0 && f.Harm2Amp > 0 {
		base += f.Harm2Amp * breath * math.Cos(float64(f.Harm2Freq)*theta+deg2rad(f.Harm2Drift)*u+deg2rad(f.Harm2Phase))
	}
	if f.RingFreq > 0 && f.RingAmp > 0 {
		base += f.RingAmp * math.Cos(2*math.Pi*f.RingFreq*u+deg2rad(f.RingDrift)*u)
	}
	return base
}

// Evaluate implements sdf.SDF3. It returns negative values inside the
// material and positive values outside. The shade is a thin shell with a
// closed bottom plate (with a central hole) and an open top rim. The
// radial-distance approximation keeps the evaluation cheap; small
// deviations on steep slopes (the flowing fluid pattern) are acceptable
// for a decorative lampshade.
func (s *fluidSDF) Evaluate(p v3.Vec) float64 {
	r := math.Hypot(p.X, p.Y)
	theta := math.Atan2(p.Y, p.X)
	z := p.Z
	wall := s.f.WallThickness

	// --- Shell wall: signed radial distance to the wall band
	// (the wall spans [outer-wall, outer] around the patterned surface). ---
	u := math.Min(math.Max(z/s.f.ShadeHeight, 0), 1)
	outer := s.surfaceRadius(u, theta)
	dWall := math.Abs(r-(outer-wall/2)) - wall/2

	// --- Bottom plate: flat disc from the wall in to a central hole. ---
	innerR := s.surfaceRadius(0, theta) - wall
	holeR := s.f.BottomHoleDiameter / 2
	plateT := wall
	dPlateR := math.Max(holeR-r, r-innerR)     // radial distance to the plate band
	dPlateZ := math.Abs(z-plateT/2) - plateT/2 // axial distance to the plate slab
	dPlate := math.Max(dPlateR, dPlateZ)       // plate = band intersected with slab

	d := math.Min(dWall, dPlate)

	// --- Ends: open top rim, closed bottom plate. ---
	if z > s.f.ShadeHeight {
		// Above the top: nearest material is the top rim annulus.
		return math.Hypot(z-s.f.ShadeHeight, math.Max(dWall, 0))
	}
	if z < 0 {
		// Below the bottom: nearest material is the plate disc.
		return math.Hypot(-z, math.Max(dPlateR, 0))
	}
	return d
}

// BoundingBox implements sdf.SDF3. The shade spans z in [0, ShadeHeight].
func (s *fluidSDF) BoundingBox() sdf.Box3 {
	return sdf.NewBox3(
		v3.Vec{X: 0, Y: 0, Z: s.f.ShadeHeight / 2},
		v3.Vec{X: 2 * s.maxRadius, Y: 2 * s.maxRadius, Z: s.f.ShadeHeight},
	)
}
