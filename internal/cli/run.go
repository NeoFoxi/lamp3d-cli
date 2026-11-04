// Run contains the CLI orchestration of the lamp generator: flag parsing,
// generator/mode/seed resolution, design randomization and STL rendering.
// It is shared by the entry points in the repository root and
// cmd/lamp3d-cli.
package cli

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/deadsy/sdfx/sdf"
	"github.com/deadsy/sdfx/vec/v3"

	"lamp3d-cli/internal/design"
	"lamp3d-cli/internal/export"
	"lamp3d-cli/internal/fluid"
	"lamp3d-cli/internal/geometry"
)

// buildDir is the folder where all generated STL files are placed. It is
// created at startup when missing.
const buildDir = "build"

// The selectable generators.
const (
	genClassic = "classic"
	genFluid   = "fluid"
	genAny     = "any"
)

// Run executes the generator with the given command line arguments.
func Run() {
	seed := flag.Int64("seed", 0, "random seed for the lamp design (0 = ask, empty answer = random from clock)")
	generator := flag.String("generator", "", "generator: classic, fluid or any (empty = interactive picker)")
	mode := flag.String("mode", "", "shade style, depends on -generator: classic: wave, fluted, spiral, ripple, dune, any; fluid: flow, bloom, vortex, weave, ripple, any (empty = interactive picker)")
	out := flag.String("out", "", "output file name prefix (default: <seed>-<mode>, fluid: <seed>-fluid-<mode>)")
	res := flag.Int("res", 300, "marching cubes resolution (higher = smoother, slower)")
	hole := flag.Float64("hole", 30, "diameter of the hole in the shade's bottom plate (mm)")
	flag.Parse()

	if *hole <= 0 {
		log.Fatalf("Hole diameter must be positive, got %v", *hole)
	}

	// Resolve the generator: flag value or interactive picker (up/down keys).
	gen := *generator
	if gen == "" {
		gen = SelectGenerator()
	}
	if gen != genClassic && gen != genFluid && gen != genAny {
		log.Fatalf("Unknown generator %q (allowed: classic, fluid, any)", gen)
	}

	// Resolve the mode: flag value or interactive picker; the valid set
	// depends on the chosen generator.
	m := design.DesignMode(*mode)
	switch gen {
	case genAny:
		if m != "" && m != design.ModeAny {
			log.Fatalf("Mode %q cannot be forced with -generator any; pick a concrete generator (-generator classic or fluid) to force a mode", *mode)
		}
		m = ""
	case genClassic:
		if m == "" {
			m = SelectMode()
		}
		if !design.ValidMode(m) {
			log.Fatalf("Unknown mode %q for the classic generator (allowed: wave, fluted, spiral, ripple, dune, any; the fluid generator allows: flow, bloom, vortex, weave, ripple, any)", *mode)
		}
	case genFluid:
		if m == "" {
			m = SelectFluidMode()
		}
		if !fluid.ValidMode(m) {
			log.Fatalf("Unknown mode %q for the fluid generator (allowed: flow, bloom, vortex, weave, ripple, any; the classic generator allows: wave, fluted, spiral, ripple, dune, any)", *mode)
		}
	}

	// Resolve the seed: flag value, interactive answer, or clock.
	*seed = ResolveSeed(*seed)
	log.Printf("Using seed: %d", *seed)

	// One rand source from the seed drives both the generator draw and the
	// design randomization, keeping every output reproducible.
	rng := rand.New(rand.NewSource(*seed))

	if gen == genAny {
		// Pick the generator at random and leave the style choice to the
		// generator's own randomizer.
		if rng.Intn(2) == 0 {
			gen = genClassic
		} else {
			gen = genFluid
		}
		m = design.ModeAny
	}
	log.Printf("Generator: %s", gen)

	// Output folder: created upfront when missing.
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		log.Fatalf("Error creating build folder: %v", err)
	}

	switch gen {
	case genClassic:
		// Build the design: fixed hardware dimensions, randomized shape.
		p := design.DefaultParams()
		p.Mode = m
		design.RandomizeParams(rng, p)
		logDesign(p)
		p.RenderRes = *res
		p.BottomHoleDiameter = *hole
		prefix := *out
		if prefix == "" {
			prefix = fmt.Sprintf("%d-%s", *seed, p.Mode)
		}
		runClassic(p, uniqueOutPath(prefix))
	case genFluid:
		// Build the design: fixed hardware dimensions, randomized shape.
		f := fluid.DefaultFluidParams()
		f.Mode = m
		fluid.RandomizeParams(rng, f)
		logFluidDesign(f)
		f.RenderRes = *res
		f.BottomHoleDiameter = *hole
		prefix := *out
		if prefix == "" {
			prefix = fmt.Sprintf("%d-fluid-%s", *seed, f.Mode)
		}
		runFluid(f, uniqueOutPath(prefix))
	}
}

// uniqueOutPath returns a non-colliding output path for the given prefix
// inside the build folder.
func uniqueOutPath(prefix string) string {
	return export.UniquePath(filepath.Join(buildDir, prefix+".stl"))
}

// runClassic builds and renders the classic lamp: base + shade merged as a
// single printable model.
func runClassic(p *design.Params, outPath string) {
	log.Println("Building lampshade ...")
	shade, err := geometry.NewShade(p)
	if err != nil {
		log.Fatalf("Error building lampshade: %v", err)
	}
	log.Println("Building base ...")
	base, err := geometry.NewBase(p)
	if err != nil {
		log.Fatalf("Error building base: %v", err)
	}
	// Place the shade on top of the base and merge both parts.
	shade = sdf.Transform3D(shade, sdf.Translate3d(v3.Vec{Z: p.BaseHeight}))
	lamp := sdf.Union3D(base, shade)
	log.Println("Writing printable model ...")
	export.WriteSTL(lamp, outPath, p.RenderRes)
	printSummary(outPath, p.PuckDiameter, p.PuckHeight)
}

// runFluid builds and renders the fluid lamp: base + shade merged as a
// single printable model.
func runFluid(f *fluid.Params, outPath string) {
	log.Println("Building lampshade ...")
	shade, err := geometry.NewFluidShade(f)
	if err != nil {
		log.Fatalf("Error building lampshade: %v", err)
	}
	log.Println("Building base ...")
	base, err := geometry.NewBase(&f.Params)
	if err != nil {
		log.Fatalf("Error building base: %v", err)
	}
	// Place the shade on top of the base and merge both parts.
	shade = sdf.Transform3D(shade, sdf.Translate3d(v3.Vec{Z: f.BaseHeight}))
	lamp := sdf.Union3D(base, shade)
	log.Println("Writing printable model ...")
	export.WriteSTL(lamp, outPath, f.RenderRes)
	printSummary(outPath, f.PuckDiameter, f.PuckHeight)
}

// printSummary logs the generated file and the print hints shared by both
// generators.
func printSummary(outPath string, puckDiameter, puckHeight float64) {
	fmt.Println()
	log.Println("Done! Generated file:")
	printFileInfo(outPath)
	fmt.Println()
	log.Println("Print hints:")
	log.Println("  Translucent PLA/PETG so the light shines through; 0.4 mm nozzle.")
	log.Println("  The bottom plate closes the shade; the LED unit lights through the hole.")
	log.Println("Assembly:")
	log.Printf("  Glue the LED unit (Ø%.0f x %.0f mm) into the base pocket with the", puckDiameter, puckHeight)
	log.Println("  included double-sided tape, LED facing up, then place the lamp upright.")
}

// logDesign prints the chosen classic style and shape values so a design
// can be reproduced or manually tweaked in internal/design.
func logDesign(p *design.Params) {
	log.Printf("Mode: %s", p.Mode)
	log.Printf("Design: height=%.0f radius=%.1f waves=%.1f/%.1f ribs=%d ribDepth=%.1f twist=%.0f° wall=%.1f",
		p.ShadeHeight, p.ShadeRadius, p.WaveAmp1, p.WaveAmp2,
		p.RibCount, p.RibAmplitude, p.TwistDegrees, p.WallThickness)
}

// logFluidDesign prints the chosen fluid style and pattern values so a
// design can be reproduced or manually tweaked in internal/fluid.
func logFluidDesign(f *fluid.Params) {
	log.Printf("Mode: %s", f.Mode)
	log.Printf("Design: height=%.0f radius=%.1f bulge=%.1f@%.0f° wall=%.1f",
		f.ShadeHeight, f.ShadeRadius, f.BulgeAmp, f.BulgePhase, f.WallThickness)
	log.Printf("Harmonics: harm1=%d/%.1f drift=%.0f° harm2=%d/%.1f drift=%.0f°",
		f.Harm1Freq, f.Harm1Amp, f.Harm1Drift, f.Harm2Freq, f.Harm2Amp, f.Harm2Drift)
	log.Printf("Rings: freq=%.1f amp=%.1f drift=%.0f° breath: freq=%.1f depth=%.2f",
		f.RingFreq, f.RingAmp, f.RingDrift, f.BreathFreq, f.BreathDepth)
}

// printFileInfo logs the size of a generated file.
func printFileInfo(path string) {
	info, err := os.Stat(path)
	if err != nil {
		log.Printf("  %s: missing (%v)", path, err)
		return
	}
	log.Printf("  %s: %.1f MB", path, float64(info.Size())/(1024*1024))
}
