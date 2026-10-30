// Package design holds the parametric lamp design: all dimensions, the
// design modes, the silhouette math and the seeded randomization.
//
// It is intentionally free of geometry/rendering dependencies so it can be
// reused by other generators (e.g. a web configurator or a different CAD
// kernel). Callers can dot-import this package to use it unqualified.
package design

import "math"

// DesignMode selects the visual style of the lampshade.
type DesignMode string

const (
	ModeAny    DesignMode = "any"    // pick a random mode (CLI default)
	ModeWave   DesignMode = "wave"   // fine spiral ribs on a wavy silhouette
	ModeFluted DesignMode = "fluted" // few deep flutes, gently twisted
	ModeSpiral DesignMode = "spiral" // one to three bold helical ribs
	ModeRipple DesignMode = "ripple" // horizontal rings, no twist
	ModeDune   DesignMode = "dune"   // wavy silhouette plus vertical lobes
)

// allModes lists the modes the random generator can pick from.
var allModes = []DesignMode{ModeWave, ModeFluted, ModeSpiral, ModeRipple, ModeDune}

// Modes returns the selectable design modes (without the "any" wildcard).
func Modes() []DesignMode {
	return append([]DesignMode(nil), allModes...)
}

// ValidMode reports whether the given design mode is known.
func ValidMode(m DesignMode) bool {
	if m == ModeAny {
		return true
	}
	for _, known := range allModes {
		if m == known {
			return true
		}
	}
	return false
}

// degToRad converts degrees to radians.
func degToRad(deg float64) float64 { return deg * math.Pi / 180 }

// Params holds every dimension of the lamp. All values are in millimeters
// (angles in degrees) so the whole design can be tuned from one place.
// The hardware dimensions of the LED unit stay fixed; everything else can
// be randomized from a seed.
type Params struct {
	// --- Style ---
	Mode DesignMode // which shade style to generate

	// --- Lampshade ---
	ShadeHeight   float64 // total height of the lampshade
	ShadeRadius   float64 // average silhouette radius of the shade
	WaveAmp1      float64 // amplitude of the main silhouette wave (the large bulges)
	WaveAmp2      float64 // amplitude of the secondary wave (adds organic variation)
	WavePhase1    float64 // phase of the main wave in degrees
	WavePhase2    float64 // phase of the secondary wave in degrees
	RibCount      int     // number of vertical ribs around the circumference
	RibAmplitude  float64 // depth of each rib or ring
	TwistDegrees  float64 // total twist of the rib spiral over the shade height (sign = direction)
	WallThickness float64 // shell wall thickness of the shade

	// --- Mode-specific knobs (zero when unused) ---
	SpiralArms    int     // ModeSpiral: number of helical ribs
	RingCount     int     // ModeRipple: number of horizontal rings
	DuneLobes     int     // ModeDune: number of vertical lobes
	DuneAmplitude float64 // ModeDune: depth of the vertical lobes

	// --- Closed bottom plate ---
	BottomHoleDiameter float64 // diameter of the hole in the shade's bottom plate

	// --- LED unit (hardware: round puck, 80 mm diameter, 20 mm high) ---
	PuckDiameter float64 // diameter of the round LED unit
	PuckHeight   float64 // height of the round LED unit

	// --- Lamp base (universal: one base fits every generated shade) ---
	BaseRadius        float64 // radius of the round base
	BaseHeight        float64 // total height of the base
	MaxBottomOuter    float64 // max allowed outer radius of a shade bottom rim
	PocketClearance   float64 // extra diameter so the LED unit drops in freely
	PocketDepthOffset float64 // how far the LED unit sits below the base top surface
	CableSlotWidth    float64 // width of the side slot for the USB cable

	// --- Rendering ---
	RenderRes int // marching cubes resolution (higher = smoother, slower)
}

// DefaultParams returns the deterministic default design (ModeWave). It is
// also the fallback if the random generator cannot produce a valid variant.
func DefaultParams() *Params {
	return &Params{
		// Style
		Mode: ModeWave,

		// Lampshade
		ShadeHeight:   200,
		ShadeRadius:   60,
		WaveAmp1:      12,
		WaveAmp2:      4,
		WavePhase1:    -17.2,
		WavePhase2:    57.3,
		RibCount:      48,
		RibAmplitude:  2.2,
		TwistDegrees:  110,
		WallThickness: 2.0,

		// Closed bottom plate
		BottomHoleDiameter: 30,

		// LED unit
		PuckDiameter: 80,
		PuckHeight:   20,

		// Base (fixed so a single base works with every design)
		BaseRadius:        82,
		BaseHeight:        25,
		MaxBottomOuter:    80,
		PocketClearance:   0.8,
		PocketDepthOffset: 0.6,
		CableSlotWidth:    4.0,

		// Rendering
		RenderRes: 300,
	}
}

// SilhouetteRadius returns the average outer radius of the shade at
// normalized height u (0 = bottom, 1 = top). Two overlapping sine waves
// create the organic wavy outline seen in the reference photo.
func SilhouetteRadius(p *Params, u float64) float64 {
	return p.ShadeRadius +
		p.WaveAmp1*math.Sin(2*math.Pi*u+degToRad(p.WavePhase1)) +
		p.WaveAmp2*math.Sin(2*math.Pi*2.3*u+degToRad(p.WavePhase2))
}
