// Run contains the CLI orchestration of the lamp generator: flag parsing,
// mode/seed resolution, design randomization and STL rendering. It is shared
// by the entry points in the repository root and cmd/lamp3d-cli.
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
	"lamp3d-cli/internal/geometry"
)

// buildDir is the folder where all generated STL files are placed. It is
// created at startup when missing.
const buildDir = "build"

// Run executes the generator with the given command line arguments.
func Run() {
	seed := flag.Int64("seed", 0, "random seed for the lamp design (0 = ask, empty answer = random from clock)")
	mode := flag.String("mode", "", "shade style: wave, fluted, spiral, ripple, dune or any (empty = interactive picker)")
	out := flag.String("out", "", "output file name prefix (default: <seed>-<mode>)")
	res := flag.Int("res", 300, "marching cubes resolution (higher = smoother, slower)")
	hole := flag.Float64("hole", 30, "diameter of the hole in the shade's bottom plate (mm)")
	flag.Parse()

	if *hole <= 0 {
		log.Fatalf("Hole diameter must be positive, got %v", *hole)
	}

	// Resolve the mode: flag value or interactive picker (up/down keys).
	p := design.DefaultParams()
	p.Mode = design.DesignMode(*mode)
	if p.Mode == "" {
		p.Mode = SelectMode()
	}
	if !design.ValidMode(p.Mode) {
		log.Fatalf("Unknown mode %q (allowed: wave, fluted, spiral, ripple, dune, any)", *mode)
	}

	// Resolve the seed: flag value, interactive answer, or clock.
	*seed = ResolveSeed(*seed)
	log.Printf("Using seed: %d", *seed)

	// Build the design: fixed hardware dimensions, randomized shape.
	design.RandomizeParams(rand.New(rand.NewSource(*seed)), p)
	logDesign(p)
	p.RenderRes = *res
	p.BottomHoleDiameter = *hole

	// Output folder: created upfront when missing.
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		log.Fatalf("Error creating build folder: %v", err)
	}

	// Output name: default includes seed and mode for reproducibility.
	prefix := *out
	if prefix == "" {
		prefix = fmt.Sprintf("%d-%s", *seed, p.Mode)
	}
	outPath := export.UniquePath(filepath.Join(buildDir, prefix+".stl"))

	// --- Complete lamp: base + shade as a single printable model ---
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

	// --- Summary ---
	fmt.Println()
	log.Println("Done! Generated file:")
	printFileInfo(outPath)
	fmt.Println()
	log.Println("Print hints:")
	log.Println("  Translucent PLA/PETG so the light shines through; 0.4 mm nozzle.")
	log.Println("  The bottom plate closes the shade; the LED unit lights through the hole.")
	log.Println("Assembly:")
	log.Printf("  Glue the LED unit (Ø%.0f x %.0f mm) into the base pocket with the", p.PuckDiameter, p.PuckHeight)
	log.Println("  included double-sided tape, LED facing up, then place the lamp upright.")
}

// logDesign prints the chosen style and shape values so a design can be
// reproduced or manually tweaked in internal/design.
func logDesign(p *design.Params) {
	log.Printf("Mode: %s", p.Mode)
	log.Printf("Design: height=%.0f radius=%.1f waves=%.1f/%.1f ribs=%d ribDepth=%.1f twist=%.0f° wall=%.1f",
		p.ShadeHeight, p.ShadeRadius, p.WaveAmp1, p.WaveAmp2,
		p.RibCount, p.RibAmplitude, p.TwistDegrees, p.WallThickness)
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
