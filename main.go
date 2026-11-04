// Command lamp3d-cli generates printable STL models for a seeded, 3D
// printable LED lamp: a decorative lampshade with a closed bottom plate and
// a round base that holds a round LED unit (80 mm diameter, 20 mm high)
// with a USB cable slot. Two generators are available:
//
//	classic  twisted wave-lamp styles: wave, fluted, spiral, ripple, dune
//	fluid    smooth flowing patterns: flow, bloom, vortex, weave, ripple
//
// A seed drives a random but reproducible lamp design, a mode selects the
// visual style, and -generator picks the generator:
//
//	lamp3d-cli -seed 42                            random style for seed 42
//	lamp3d-cli --seed 42 -generator fluid -mode vortex
//	lamp3d-cli --seed 42 -mode spiral              classic style
//	lamp3d-cli                                     interactive: pick generator,
//	                                               then style (up/down) and seed
//	lamp3d-cli -seed 0                             random seed from the clock
//
// The complete lamp (base + shade) is written as a single printable STL
// file per seed/generator/mode into the build folder: <seed>-<mode>.stl for
// the classic generator and <seed>-fluid-<mode>.stl for the fluid one. The
// project is organized the standard Go way:
//
//	.                    this entry point (go run .)
//	cmd/lamp3d-cli       named binary entry point
//	internal/cli         CLI orchestration and interactive terminal prompts
//	internal/design      classic parametric design: dimensions, modes, randomization
//	internal/fluid       fluid parametric design: flowing pattern knobs
//	internal/geometry    SDF builders for shade and base
//	internal/export      STL writing and unique file names
package main

import "lamp3d-cli/internal/cli"

func main() {
	cli.Run()
}
