# lamp-3d

A generator for a 3D-printable twisted "wave lamp": a ribbed, wavy
lampshade (printed in spiral/vase mode) and a round base that holds a round
LED unit (80 mm diameter, 20 mm high) with a USB cable slot — inspired by
wave-style lamps on MakerWorld.

The lamp is seeded and randomized: every seed produces a different but
reproducible design, with multiple visual styles to choose from.

## Requirements

- Go 1.21+ (module `lamp-3d`, uses [github.com/deadsy/sdfx](https://github.com/deadsy/sdfx))
- A slicer for the generated STL files (e.g. Bambu Studio, OrcaSlicer)
- Hardware: a round LED unit, Ø80 mm × 20 mm (e.g. a USB puck light)

## Build

```sh
go build -o lamp3d-cli ./cmd/lamp3d-cli
```

## Usage

```sh
./lamp3d-cli -seed 42                  # deterministic design for seed 42
./lamp3d-cli --seed 42 -mode spiral    # force a style (double dash also works)
./lamp3d-cli                           # interactive: pick style (up/down keys)
                                       # and seed (enter = random)
./lamp3d-cli -seed 0                   # random seed from the clock
```

Running the same design twice never overwrites: if the output file already
exists, a counter is appended (`42-wave.stl`, `42-wave-2.stl`, ...).

Flags:

| Flag    | Default      | Description                                          |
|---------|--------------|------------------------------------------------------|
| `-seed` | `0`          | Random seed; `0` asks interactively (enter = random) |
| `-mode` | *(ask)*      | Shade style, see below; empty shows an interactive picker with up/down keys |
| `-out`  | `lamp-<seed>-<mode>` | Output file prefix                           |
| `-res`  | `300`        | Marching cubes resolution (higher = smoother, slower)|
| `-hole` | `30`         | Diameter of the hole in the shade's bottom plate (mm)|

### Design modes

| Mode     | Look                                              |
|----------|---------------------------------------------------|
| `wave`   | Fine spiral ribs on a wavy silhouette (reference) |
| `fluted` | Few deep flutes, gently twisted                   |
| `spiral` | One to three bold helical ribs                    |
| `ripple` | Horizontal rings, no twist                        |
| `dune`   | Wavy silhouette plus large vertical lobes         |
| `any`    | Picks one of the above at random                  |

Same seed + same mode = byte-identical STL files.

## Output

The complete lamp (base + shade merged) is written as a **single printable
STL** into the `build/` folder (created automatically):

- `build/<seed>-<mode>.stl`

## Print & assembly

- Translucent PLA/PETG so the light shines through, 0.4 mm nozzle
- The bottom plate closes the shade; the LED unit lights through the hole
- Glue the LED unit into the base pocket (LED facing up, e.g. with the
  included double-sided tape); the USB cable exits through the side slot

## Project layout

Standard Go layout:

```
cmd/lamp3d-cli/       CLI entry point (flags, prompts, orchestration)
internal/design/      parametric design: dimensions, modes, randomization
internal/geometry/    SDF builders for lampshade and base
internal/export/      STL writing and unique file names
internal/cli/         interactive terminal prompts
```

Tweak anything in `internal/design/params.go` or the ranges in
`internal/design/randomize.go` and rerun.

## Disclaimer

**This project was built by AI** (generated with an AI coding assistant).
It has not been reviewed by a human engineer. Verify the generated models
in your slicer before printing, and use at your own risk.
