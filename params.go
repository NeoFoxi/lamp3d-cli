package main

// Params holds every dimension of the lamp. All values are in millimeters
// (angles in degrees) so the whole design can be tuned from one place.
// The hardware dimensions of the LED unit stay fixed; everything else can
// be randomized from a seed.
type Params struct {
	// --- Lampshade (printed in spiral/vase mode) ---
	ShadeHeight   float64 // total height of the lampshade
	ShadeRadius   float64 // average silhouette radius of the shade
	WaveAmp1      float64 // amplitude of the main silhouette wave (the large bulges)
	WaveAmp2      float64 // amplitude of the secondary wave (adds organic variation)
	WavePhase1    float64 // phase of the main wave in degrees
	WavePhase2    float64 // phase of the secondary wave in degrees
	RibCount      int     // number of vertical ribs around the circumference
	RibAmplitude  float64 // depth of each rib
	TwistDegrees  float64 // total twist of the rib spiral over the shade height (sign = direction)
	WallThickness float64 // shell wall thickness of the shade

	// --- LED unit (hardware: round puck, 80 mm diameter, 20 mm high) ---
	PuckDiameter float64 // diameter of the round LED unit
	PuckHeight   float64 // height of the round LED unit

	// --- Lamp base ---
	BaseRadius        float64 // radius of the round base
	BaseHeight        float64 // total height of the base
	PocketClearance   float64 // extra diameter so the LED unit drops in freely
	PocketDepthOffset float64 // how far the LED unit sits below the base top surface
	CableSlotWidth    float64 // width of the side slot for the USB cable
	SeatClearance     float64 // diametral clearance of the groove that aligns the shade

	// --- Rendering ---
	RenderRes int // marching cubes resolution (higher = smoother, slower)
}

// DefaultParams returns the deterministic default design. It is also the
// fallback if the random generator cannot produce a valid variant.
func DefaultParams() *Params {
	return &Params{
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

		// LED unit
		PuckDiameter: 80,
		PuckHeight:   20,

		// Base
		BaseRadius:        70,
		BaseHeight:        25,
		PocketClearance:   0.8,
		PocketDepthOffset: 0.6,
		CableSlotWidth:    4.0,
		SeatClearance:     0.4,

		// Rendering
		RenderRes: 300,
	}
}
