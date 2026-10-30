package main

import (
	"math"
	"math/rand"
)

// RandomizeParams derives a random lamp design from the given RNG. The LED
// unit dimensions and the base mechanics stay fixed; only the visual shape
// (waves, ribs, twist, proportions) is randomized. The result is validated
// so every seed produces a printable lamp that fits the hardware; if no
// valid variant is found the deterministic default design is used.
func RandomizeParams(rng *rand.Rand, p *Params) {
	for attempt := 0; attempt < 100; attempt++ {
		t := *p // start from the defaults; hardware values are kept

		// Shade shape
		t.ShadeHeight = randRange(rng, 180, 220)
		t.ShadeRadius = randRange(rng, 55, 65)
		t.WaveAmp1 = randRange(rng, 8, 14)
		t.WaveAmp2 = randRange(rng, 2, 5)
		t.WavePhase1 = rng.Float64() * 360
		t.WavePhase2 = rng.Float64() * 360
		t.RibCount = 32 + rng.Intn(33) // 32..64 ribs
		t.RibAmplitude = randRange(rng, 1.6, 2.6)
		t.TwistDegrees = randRange(rng, 60, 160)
		if rng.Intn(2) == 0 {
			t.TwistDegrees = -t.TwistDegrees // random spiral direction
		}
		t.WallThickness = randRange(rng, 1.6, 2.2)

		// Base: large enough for the shade bottom rim, comfortable margin.
		bottom := SilhouetteRadius(&t, 0)
		t.BaseRadius = math.Max(66, bottom+t.RibAmplitude+t.WallThickness+6)
		t.BaseHeight = randRange(rng, 22, 28)

		if validDesign(&t) {
			*p = t
			return
		}
	}
	// No valid random variant found: keep the deterministic default.
	*p = *DefaultParams()
}

// validDesign checks the physical constraints of a candidate design:
// the shade must clear the LED pocket at the bottom, fit onto the base,
// and the base must keep enough floor material under the LED unit.
func validDesign(p *Params) bool {
	pocketR := (p.PuckDiameter + p.PocketClearance) / 2

	// Inner wall radius of the shade at its bottom must clear the pocket.
	bottomInner := SilhouetteRadius(p, 0) - p.RibAmplitude - p.WallThickness
	if bottomInner < pocketR+3 {
		return false
	}

	// Shade bottom rim must sit on the base, not hang over its edge.
	bottomOuter := SilhouetteRadius(p, 0) + p.RibAmplitude + p.WallThickness
	if bottomOuter > p.BaseRadius-2 {
		return false
	}

	// Keep enough material under the LED pocket for the cable slot.
	floor := p.BaseHeight - p.PuckHeight - p.PocketDepthOffset
	if floor < 4 {
		return false
	}

	// Keep the shade within a sane envelope for desktop printing.
	if SilhouetteRadius(p, 1)+p.RibAmplitude+p.WallThickness > 100 {
		return false
	}
	return true
}

// randRange returns a uniform random float in [min, max].
func randRange(rng *rand.Rand, min, max float64) float64 {
	return min + rng.Float64()*(max-min)
}
