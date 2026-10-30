package design

import "math/rand"

// RandomizeParams derives a random lamp design from the given RNG. The LED
// unit dimensions and the base mechanics stay fixed; only the visual shape
// is randomized. One of the design modes is chosen (or kept when the user
// forced one), and each candidate is validated so every seed produces a
// printable lamp that fits the hardware. If no valid variant is found the
// deterministic default design is used.
func RandomizeParams(rng *rand.Rand, p *Params) {
	for attempt := 0; attempt < 100; attempt++ {
		t := *p // start from the current params; hardware values are kept

		// Common shade shape
		t.ShadeHeight = randRange(rng, 180, 220)
		t.ShadeRadius = randRange(rng, 55, 65)
		t.WaveAmp1 = randRange(rng, 8, 12)
		t.WaveAmp2 = randRange(rng, 2, 5)
		t.WavePhase1 = rng.Float64() * 360
		t.WavePhase2 = rng.Float64() * 360
		t.WallThickness = randRange(rng, 1.6, 2.2)
		t.SpiralArms, t.RingCount, t.DuneLobes, t.DuneAmplitude = 0, 0, 0, 0

		// Mode: keep a user-forced mode, otherwise pick one at random.
		if t.Mode == "" || t.Mode == ModeAny {
			t.Mode = allModes[rng.Intn(len(allModes))]
		}
		randomizeMode(rng, &t)

		if validDesign(&t) {
			*p = t
			return
		}
	}
	// No valid random variant found: keep the deterministic default.
	*p = *DefaultParams()
}

// randomizeMode fills in the mode-specific shape parameters.
func randomizeMode(rng *rand.Rand, t *Params) {
	switch t.Mode {
	case ModeSpiral:
		t.SpiralArms = 1 + rng.Intn(3)            // 1..3 helical ribs
		t.RibAmplitude = randRange(rng, 3, 5)     // bold ribs
		t.TwistDegrees = randRange(rng, 180, 360) // strong helix
	case ModeRipple:
		t.RingCount = 18 + rng.Intn(19) // 18..36 horizontal rings
		t.RibAmplitude = randRange(rng, 1.5, 3)
		t.TwistDegrees = 0 // rings stay level
	case ModeFluted:
		t.RibCount = 12 + rng.Intn(9) // 12..20 deep flutes
		t.RibAmplitude = randRange(rng, 3, 5)
		t.TwistDegrees = randRange(rng, 20, 70) // gentle twist
	case ModeDune:
		t.DuneLobes = 3 + rng.Intn(3) // 3..5 vertical lobes
		t.DuneAmplitude = randRange(rng, 3, 6)
		t.RibCount = 40 + rng.Intn(21) // fine surface texture
		t.RibAmplitude = randRange(rng, 1.2, 2)
		t.TwistDegrees = randRange(rng, 40, 120)
	default: // ModeWave
		t.RibCount = 36 + rng.Intn(25) // 36..60 fine ribs
		t.RibAmplitude = randRange(rng, 1.6, 2.6)
		t.TwistDegrees = randRange(rng, 60, 160)
	}
	// Random spiral direction for all twisted styles.
	if t.TwistDegrees != 0 && rng.Intn(2) == 0 {
		t.TwistDegrees = -t.TwistDegrees
	}
}

// validDesign checks the physical constraints of a candidate design:
// the shade must clear the LED pocket at the bottom, fit onto the base,
// and the base must keep enough floor material under the LED unit.
func validDesign(p *Params) bool {
	pocketR := (p.PuckDiameter + p.PocketClearance) / 2

	// Inner wall radius of the shade at its bottom must clear the pocket.
	bottomInner := SilhouetteRadius(p, 0) - p.RibAmplitude - p.DuneAmplitude - p.WallThickness
	if bottomInner < pocketR+3 {
		return false
	}

	// Shade bottom rim must sit on the universal base, not hang over its edge.
	bottomOuter := SilhouetteRadius(p, 0) + p.RibAmplitude + p.DuneAmplitude + p.WallThickness
	if bottomOuter > p.MaxBottomOuter {
		return false
	}

	// Keep enough material under the LED pocket for the cable slot.
	floor := p.BaseHeight - p.PuckHeight - p.PocketDepthOffset
	if floor < 4 {
		return false
	}

	// Keep the shade within a sane envelope for desktop printing.
	if SilhouetteRadius(p, 1)+p.RibAmplitude+p.DuneAmplitude+p.WallThickness > 100 {
		return false
	}
	return true
}

// randRange returns a uniform random float in [min, max].
func randRange(rng *rand.Rand, min, max float64) float64 {
	return min + rng.Float64()*(max-min)
}
