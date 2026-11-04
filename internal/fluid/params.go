// Package fluid is a second lamp generator with smooth, fluid-flowing
// surface patterns: a simple silhouette plus layered twisted angular
// harmonics, breathing amplitude modulation and drifting rings.
//
// It embeds design.Params so the shared lamp dimensions (shade, LED unit,
// base mechanics) stay identical to the classic generator, and reuses
// design.DesignMode as its mode type.
package fluid

import (
	"math"

	"lamp3d-cli/internal/design"
)

// The fluid design modes select the flowing surface pattern style.
const (
	ModeFlow   design.DesignMode = "flow"   // two twisted harmonics drifting opposite, gently breathing — fluid interference
	ModeBloom  design.DesignMode = "bloom"  // few soft lobes that swell and fade along the height — breathing petals
	ModeVortex design.DesignMode = "vortex" // one strong slow swirl plus a faint helical ring
	ModeWeave  design.DesignMode = "weave"  // two finer harmonics crossing at different twist speeds — flowing moiré
	ModeRipple design.DesignMode = "ripple" // smooth horizontal rings, optionally slightly helical
)

// fluidModes lists the modes the random generator can pick from.
var fluidModes = []design.DesignMode{ModeFlow, ModeBloom, ModeVortex, ModeWeave, ModeRipple}

// Modes returns the selectable fluid modes (without the "any" wildcard).
func Modes() []design.DesignMode {
	return append([]design.DesignMode(nil), fluidModes...)
}

// ValidMode reports whether the given design mode is known to the fluid
// generator (the "any" wildcard is accepted everywhere).
func ValidMode(m design.DesignMode) bool {
	if m == design.ModeAny {
		return true
	}
	for _, known := range fluidModes {
		if m == known {
			return true
		}
	}
	return false
}

// degToRad converts degrees to radians.
func degToRad(deg float64) float64 { return deg * math.Pi / 180 }

// Params holds every dimension of the fluid lamp. All values are in
// millimeters (angles in degrees). The shared lamp dimensions from
// design.Params are embedded; the fluid pattern knobs sit on top.
type Params struct {
	design.Params // shared shade, hardware, base and rendering dimensions

	// --- Silhouette ---
	BulgeAmp   float64 // amplitude of the single gentle silhouette wave
	BulgePhase float64 // phase of the silhouette wave in degrees

	// --- Harmonic 1 (broad lobes) ---
	Harm1Freq  int     // angular frequency (lobes around the circumference) of harmonic 1; 0 = off
	Harm1Amp   float64 // amplitude of harmonic 1
	Harm1Drift float64 // total twist of harmonic 1 over the shade height in degrees (sign = direction)
	Harm1Phase float64 // phase of harmonic 1 in degrees

	// --- Harmonic 2 (finer, usually counter-rotating) ---
	Harm2Freq  int     // angular frequency of harmonic 2; 0 = off
	Harm2Amp   float64 // amplitude of harmonic 2
	Harm2Drift float64 // total twist of harmonic 2 over the shade height in degrees (sign = direction)
	Harm2Phase float64 // phase of harmonic 2 in degrees

	// --- Drifting rings ---
	RingFreq  float64 // number of horizontal rings over the shade height; 0 = off
	RingAmp   float64 // amplitude of the rings
	RingDrift float64 // total helical drift of the rings in degrees over the height

	// --- Breathing ---
	BreathFreq  float64 // breathing cycles of the pattern over the height; 0 = off
	BreathPhase float64 // phase of the breathing in degrees
	BreathDepth float64 // modulation depth in [0,1]: pattern amplitude factor lies in [1-BreathDepth, 1]
}

// DefaultFluidParams returns the deterministic default fluid design
// (ModeFlow). It is also the fallback if the random generator cannot
// produce a valid variant.
func DefaultFluidParams() *Params {
	f := &Params{
		Params: *design.DefaultParams(),

		// Silhouette
		BulgeAmp:   10,
		BulgePhase: -30,

		// Harmonic 1
		Harm1Freq:  8,
		Harm1Amp:   3.2,
		Harm1Drift: 160,
		Harm1Phase: 0,

		// Harmonic 2
		Harm2Freq:  18,
		Harm2Amp:   1.4,
		Harm2Drift: -80,
		Harm2Phase: 40,

		// Rings (off by default)
		RingFreq:  0,
		RingAmp:   0,
		RingDrift: 0,

		// Breathing
		BreathFreq:  1.5,
		BreathPhase: 0,
		BreathDepth: 0.25,
	}
	// Promoted fields cannot be set in the composite literal above, so
	// the fluid default mode is assigned separately.
	f.Mode = ModeFlow
	return f
}

// SilhouetteRadius returns the outer radius of the fluid shade at
// normalized height u (0 = bottom, 1 = top). A single gentle sine wave
// keeps the outline smooth; the flowing detail lives in the surface
// pattern, not the silhouette.
func SilhouetteRadius(f *Params, u float64) float64 {
	return f.ShadeRadius + f.BulgeAmp*math.Sin(2*math.Pi*u+degToRad(f.BulgePhase))
}

// PatternAmpMax returns a conservative bound on the total pattern
// amplitude, used for the physical clearance checks.
func PatternAmpMax(f *Params) float64 {
	return f.Harm1Amp + f.Harm2Amp + f.RingAmp
}
