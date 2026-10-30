// Command lamp3d generates printable STL models for a twisted wave lamp:
// a ribbed, wavy lampshade (printed in vase mode) and a round base that
// holds a round LED unit (80 mm diameter, 20 mm high) with a USB cable slot.
//
// A seed drives a random but reproducible lamp design:
//
//	lamp3d -seed 42        deterministic design for seed 42
//	lamp3d --seed 42       same (double dash also works)
//	lamp3d                 asks for a seed interactively
//	lamp3d -seed 0         picks a random seed from the clock
//
// The geometry is split into params.go, shade.go, base.go and randomize.go;
// this file only handles the CLI, the model building and the STL export.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
)

func main() {
	seed := flag.Int64("seed", 0, "random seed for the lamp design (0 = ask, empty answer = random from clock)")
	out := flag.String("out", "", "output file prefix (default: lamp-<seed>)")
	res := flag.Int("res", 300, "marching cubes resolution (higher = smoother, slower)")
	flag.Parse()

	// Resolve the seed: flag value, interactive answer, or clock.
	*seed = resolveSeed(*seed)
	log.Printf("Using seed: %d", *seed)

	// Build the design: fixed hardware dimensions, randomized shape.
	p := DefaultParams()
	RandomizeParams(rand.New(rand.NewSource(*seed)), p)
	logDesign(p)

	// Output prefix: default includes the seed for reproducibility.
	prefix := *out
	if prefix == "" {
		prefix = fmt.Sprintf("lamp-%d", *seed)
	}
	p.RenderRes = *res

	// --- Lampshade ---
	log.Println("Building lampshade ...")
	shade, err := NewShade(p)
	if err != nil {
		log.Fatalf("Error building lampshade: %v", err)
	}
	writeSTL(shade, prefix+"_shade.stl", p.RenderRes)

	// --- Base ---
	log.Println("Building base ...")
	base, err := NewBase(p)
	if err != nil {
		log.Fatalf("Error building base: %v", err)
	}
	writeSTL(base, prefix+"_base.stl", p.RenderRes)

	// --- Summary ---
	fmt.Println()
	log.Println("Done! Generated files:")
	printFileInfo(prefix + "_shade.stl")
	printFileInfo(prefix + "_base.stl")
	fmt.Println()
	log.Println("Print hints:")
	log.Println("  Shade: spiral/vase mode, white translucent PLA/PETG, 0.4 mm nozzle.")
	log.Println("  Base:  any color, 15-20% infill, no supports needed.")
	log.Println("Assembly:")
	log.Printf("  Drop the LED unit (Ø%.0f x %.0f mm) into the top pocket of the base,", p.PuckDiameter, p.PuckHeight)
	log.Println("  route the USB cable through the side slot, then place the shade over it.")
}

// resolveSeed returns the seed to use. A nonzero flag value wins. Otherwise
// the user is asked on the terminal; pressing enter there picks a random
// seed from the current time.
func resolveSeed(flagValue int64) int64 {
	if flagValue != 0 {
		return flagValue
	}
	fmt.Print("Enter a seed (press enter for a random one): ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err == nil {
		line = strings.TrimSpace(line)
		if line != "" {
			if v, err := strconv.ParseInt(line, 10, 64); err == nil {
				return v
			}
			log.Fatalf("Invalid seed %q: must be an integer", line)
		}
	}
	return time.Now().UnixNano()
}

// logDesign prints the randomized shape values so a design can be
// reproduced or manually tweaked in params.go.
func logDesign(p *Params) {
	log.Printf("Design: height=%.0f radius=%.1f waves=%.1f/%.1f ribs=%d ribDepth=%.1f twist=%.0f° wall=%.1f",
		p.ShadeHeight, p.ShadeRadius, p.WaveAmp1, p.WaveAmp2,
		p.RibCount, p.RibAmplitude, p.TwistDegrees, p.WallThickness)
}

// writeSTL renders an SDF3 to an STL file.
func writeSTL(s sdf.SDF3, path string, res int) {
	render.ToSTL(s, path, render.NewMarchingCubesUniform(res))
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
