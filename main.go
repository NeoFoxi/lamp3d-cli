// Command lamp3d-cli generates printable STL models for a twisted wave lamp:
// a ribbed, wavy lampshade with a closed bottom plate and a round base that
// holds a round LED unit (80 mm diameter, 20 mm high) with a USB cable slot.
//
// A seed drives a random but reproducible lamp design, and a mode selects
// the visual style:
//
//	lamp3d-cli -seed 42                  random style for seed 42
//	lamp3d-cli --seed 42 -mode spiral    force a style (wave, fluted, spiral,
//	                                     ripple, dune or any)
//	lamp3d-cli                           interactive: pick style (up/down) and seed
//	lamp3d-cli -seed 0                   picks a random seed from the clock
//
// The complete lamp (base + shade) is written as a single printable STL
// file per seed/mode into the build folder. The project is organized the
// standard Go way:
//
//	.                    this entry point (go run .)
//	cmd/lamp3d-cli       named binary entry point
//	internal/cli         CLI orchestration and interactive terminal prompts
//	internal/design      parametric design: dimensions, modes, randomization
//	internal/geometry    SDF builders for shade and base
//	internal/export      STL writing and unique file names
package main

import "lamp-3d/internal/cli"

func main() {
	cli.Run()
}
