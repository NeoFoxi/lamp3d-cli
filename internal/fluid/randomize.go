package fluid

import (
	"math/rand"

	"lamp3d-cli/internal/design"
)

// RandomizeParams derives a random fluid lamp design from the given RNG. The
// LED unit dimensions and the base mechanics stay fixed; only the visual
// shape is randomized. One of the fluid modes is chosen (or kept when the
// user forced one), and each candidate is validated so every seed produces
// a printable lamp that fits the hardware. If no valid variant is found the
// deterministic default fluid design is used.
func RandomizeParams(rng *rand.Rand, f *Params) {
	for attempt := 0; attempt < 100; attempt++ {
		t := *f // start from the current params; hardware values are kept

		// Common shade shape
		t.ShadeHeight = randRange(rng, 180, 220)
		t.ShadeRadius = randRange(rng, 55, 65)
		t.BulgeAmp = randRange(rng, 7, 12)
		t.BulgePhase = rng.Float64() * 360
		t.WallThickness = randRange(rng, 1.6, 2.2)

		// Reset all pattern knobs; the mode randomizer fills them in.
		t.Harm1Freq, t.Harm1Amp, t.Harm1Drift, t.Harm1Phase = 0, 0, 0, 0
		t.Harm2Freq, t.Harm2Amp, t.Harm2Drift, t.Harm2Phase = 0, 0, 0, 0
		t.RingFreq, t.RingAmp, t.RingDrift = 0, 0, 0
		t.BreathFreq, t.BreathPhase, t.BreathDepth = 0, 0, 0

		// Mode: keep a user-forced mode, otherwise pick one at random.
		if t.Mode == "" || t.Mode == design.ModeAny {
			t.Mode = fluidModes[rng.Intn(len(fluidModes))]
		}
		randomizeFluidMode(rng, &t)

		if validFluid(&t) {
			*f = t
			return
		}
	}
	// No valid random variant found: keep the deterministic default.
	*f = *DefaultFluidParams()
}

// randomizeFluidMode fills in the mode-specific pattern parameters. After
// applying a mode, each nonzero drift is flipped once, randomly and
// independently, for the spiral direction.
func randomizeFluidMode(rng *rand.Rand, t *Params) {
	switch t.Mode {
	case ModeFlow:
		// Two twisted harmonics drifting opposite, gently breathing.
		t.Harm1Freq = 6 + rng.Intn(5) // 6..10 lobes
		t.Harm1Amp = randRange(rng, 2.8, 4.2)
		t.Harm1Drift = randRange(rng, 120, 220)
		t.Harm1Phase = rng.Float64() * 360

		t.Harm2Freq = 14 + rng.Intn(9) // 14..22 finer lobes
		t.Harm2Amp = randRange(rng, 1.2, 2.2)
		t.Harm2Drift = randRange(rng, 60, 140)
		t.Harm2Phase = rng.Float64() * 360

		t.BreathFreq = randRange(rng, 1, 2)
		t.BreathDepth = randRange(rng, 0.15, 0.3)
		t.BreathPhase = rng.Float64() * 360
	case ModeBloom:
		// Few soft lobes that swell and fade along the height.
		t.Harm1Freq = 5 + rng.Intn(4) // 5..8 lobes
		t.Harm1Amp = randRange(rng, 3.5, 5.5)
		t.Harm1Drift = randRange(rng, 0, 30)

		t.BreathFreq = randRange(rng, 1.5, 3)
		t.BreathDepth = randRange(rng, 0.35, 0.55)
		t.BreathPhase = rng.Float64() * 360
	case ModeVortex:
		// One strong slow swirl plus a faint helical ring.
		t.Harm1Freq = 3 + rng.Intn(3) // 3..5 lobes
		t.Harm1Amp = randRange(rng, 2.5, 4)
		t.Harm1Drift = randRange(rng, 360, 540)

		t.RingFreq = randRange(rng, 2, 3)
		t.RingAmp = randRange(rng, 0.8, 1.5)
		t.RingDrift = randRange(rng, 120, 240)
	case ModeWeave:
		// Two finer harmonics crossing at different twist speeds.
		t.Harm1Freq = 16 + rng.Intn(9) // 16..24 lobes
		t.Harm1Amp = randRange(rng, 1.6, 2.6)
		t.Harm1Drift = randRange(rng, 80, 160)
		t.Harm1Phase = rng.Float64() * 360

		t.Harm2Freq = 9 + rng.Intn(8) // 9..16 lobes
		t.Harm2Amp = randRange(rng, 1.6, 2.6)
		t.Harm2Drift = randRange(rng, 80, 160)
		t.Harm2Phase = rng.Float64() * 360
	case ModeRipple:
		// Smooth horizontal rings, optionally slightly helical.
		t.RingFreq = randRange(rng, 14, 26)
		t.RingAmp = randRange(rng, 1.5, 2.8)
		t.RingDrift = randRange(rng, 0, 40)
	}

	// Counter-rotating harmonics: mode randomizers set both drifts with
	// the same sign above; flip harmonic 2 opposite to harmonic 1.
	if t.Harm1Drift != 0 && t.Harm2Drift != 0 {
		t.Harm2Drift = -t.Harm2Drift
	}

	// Random spiral direction for each nonzero drift, independently.
	if t.Harm1Drift != 0 && rng.Intn(2) == 0 {
		t.Harm1Drift = -t.Harm1Drift
	}
	if t.Harm2Drift != 0 && rng.Intn(2) == 0 {
		t.Harm2Drift = -t.Harm2Drift
	}
	if t.RingDrift != 0 && rng.Intn(2) == 0 {
		t.RingDrift = -t.RingDrift
	}
}

// validFluid checks the physical constraints of a candidate fluid design:
// the shade must clear the LED pocket at the bottom, fit onto the base,
// and the base must keep enough floor material under the LED unit.
func validFluid(f *Params) bool {
	pocketR := (f.PuckDiameter + f.PocketClearance) / 2
	pattern := PatternAmpMax(f)

	// Inner wall radius of the shade at its bottom must clear the pocket.
	bottomInner := SilhouetteRadius(f, 0) - pattern - f.WallThickness
	if bottomInner < pocketR+3 {
		return false
	}

	// Shade bottom rim must sit on the universal base, not hang over its edge.
	bottomOuter := SilhouetteRadius(f, 0) + pattern + f.WallThickness
	if bottomOuter > f.MaxBottomOuter {
		return false
	}

	// Keep enough material under the LED pocket for the cable slot.
	floor := f.BaseHeight - f.PuckHeight - f.PocketDepthOffset
	if floor < 4 {
		return false
	}

	// Keep the shade within a sane envelope for desktop printing.
	if SilhouetteRadius(f, 1)+pattern+f.WallThickness > 100 {
		return false
	}
	return true
}

// randRange returns a uniform random float in [min, max].
func randRange(rng *rand.Rand, min, max float64) float64 {
	return min + rng.Float64()*(max-min)
}
