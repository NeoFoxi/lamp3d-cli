# lamp3d-cli

A generator for 3D-printable seeded LED lamps: a decorative lampshade with a
closed bottom plate and a round base that holds a round LED unit (80 mm
diameter, 20 mm high) with a USB cable slot.

Two generators ship in one binary:

- **classic** — the twisted "wave lamp": ribbed, wavy shades, inspired by
  wave-style lamps on MakerWorld
- **fluid** — smooth, flowing surface patterns built from twisted
  harmonics, drifting rings and breathing amplitude

Every seed produces a different but reproducible design, with multiple
visual styles to choose from.

## Vibe coded

This project is **vibe coded**: it was written end to end by an AI coding
assistant, and no human engineer has reviewed it. The generated models are
for tinkerers — always verify the STL in your slicer before printing, and
print at your own risk.

## Requirements

- Go 1.21+ (module `lamp3d-cli`, uses [github.com/deadsy/sdfx](https://github.com/deadsy/sdfx))
- A slicer for the generated STL files (e.g. Bambu Studio, OrcaSlicer)
- Hardware: a round LED unit, Ø80 mm × 20 mm (e.g. a USB puck light)

## Build

```sh
go build -o lamp3d-cli ./cmd/lamp3d-cli
```

## Usage

```sh
./lamp3d-cli -seed 42                                # deterministic design for seed 42
./lamp3d-cli --seed 42 -generator fluid -mode vortex # fluid generator, vortex style
./lamp3d-cli --seed 42 -mode spiral                  # classic generator (double dash also works)
./lamp3d-cli -generator fluid                        # interactive seed and fluid style
./lamp3d-cli                                         # interactive: pick generator, then
                                                     # style (up/down keys) and seed
./lamp3d-cli -seed 0                                 # random seed from the clock
```

Interactive flow: first the generator, then the style (up/down keys, enter
to confirm), then the seed (enter = random). With `-generator any` no style
can be forced; both the generator and the style are then picked at random
from the seed.

Running the same design twice never overwrites: if the output file already
exists, a counter is appended (`42-wave.stl`, `42-wave-2.stl`, ...).

Flags:

| Flag         | Default                              | Description                                                                   |
|--------------|--------------------------------------|-------------------------------------------------------------------------------|
| `-seed`      | `0`                                  | Random seed; `0` asks interactively (enter = random)                          |
| `-generator` | *(ask)*                              | Generator: `classic`, `fluid` or `any` (random); empty shows an interactive picker |
| `-mode`      | *(ask)*                              | Shade style for the chosen generator, see below; empty shows an interactive picker with up/down keys |
| `-out`       | `<seed>-<mode>` / `<seed>-fluid-<mode>` | Output file prefix                                                          |
| `-res`       | `300`                                | Marching cubes resolution (higher = smoother, slower)                         |
| `-hole`      | `30`                                 | Diameter of the hole in the shade's bottom plate (mm)                         |

### Classic modes

| Mode     | Look                                              |
|----------|---------------------------------------------------|
| `wave`   | Fine spiral ribs on a wavy silhouette (reference) |
| `fluted` | Few deep flutes, gently twisted                   |
| `spiral` | One to three bold helical ribs                    |
| `ripple` | Horizontal rings, no twist                        |
| `dune`   | Wavy silhouette plus large vertical lobes         |
| `any`    | Picks one of the above at random                  |

### Fluid modes

| Mode     | Look                                                                   |
|----------|------------------------------------------------------------------------|
| `flow`   | Two twisted harmonics drifting opposite ways — fluid interference      |
| `bloom`  | Few soft lobes that swell and fade along the height — breathing petals |
| `vortex` | One strong slow swirl plus a faint helical ring                        |
| `weave`  | Two finer harmonics crossing at different twist speeds — flowing moiré |
| `ripple` | Smooth horizontal rings, optionally slightly helical                   |
| `any`    | Picks one of the above at random                                       |

Same seed + same generator + same mode = byte-identical STL files.

## Output

The complete lamp (base + shade merged) is written as a **single printable
STL** into the `build/` folder (created automatically):

- classic: `build/<seed>-<mode>.stl`
- fluid: `build/<seed>-fluid-<mode>.stl`

## Print & assembly

- Translucent PLA/PETG so the light shines through, 0.4 mm nozzle
- The bottom plate closes the shade; the LED unit lights through the hole
- Glue the LED unit into the base pocket (LED facing up, e.g. with the
  included double-sided tape); the USB cable exits through the side slot
