package geometry

import (
	"github.com/deadsy/sdfx/sdf"
	"github.com/deadsy/sdfx/vec/v3"

	"lamp3d-cli/internal/design"
)

// NewBase builds the round lamp base sized for a round LED unit
// (PuckDiameter x PuckHeight). The base has:
//   - an open pocket on top that holds the LED unit (facing up),
//   - a side slot so the USB cable can exit below the unit,
//   - a wide, shallow step on top that visually frames the shade.
func NewBase(p *design.Params) (sdf.SDF3, error) {
	// Outer body: rounded cylinder sitting on z = 0.
	body, err := sdf.Cylinder3D(p.BaseHeight, p.BaseRadius, 2.0)
	if err != nil {
		return nil, err
	}
	body = sdf.Transform3D(body, sdf.Translate3d(v3.Vec{Z: p.BaseHeight / 2}))

	// LED pocket: open from the top, deep enough for the unit plus offset.
	pocketR := (p.PuckDiameter + p.PocketClearance) / 2
	pocketDepth := p.PuckHeight + p.PocketDepthOffset
	pocket, err := sdf.Cylinder3D(pocketDepth*2, pocketR, 0)
	if err != nil {
		return nil, err
	}
	// Centered so it spans [BaseHeight - 2*depth, BaseHeight]: cuts a pocket
	// of the wanted depth out of the top face.
	pocket = sdf.Transform3D(pocket, sdf.Translate3d(v3.Vec{Z: p.BaseHeight - pocketDepth}))
	body = sdf.Difference3D(body, pocket)

	// USB cable slot from the pocket floor through the side wall.
	slot, err := cableSlot(p, pocketR)
	if err != nil {
		return nil, err
	}
	body = sdf.Difference3D(body, slot)

	// Wide, shallow alignment step around the pocket. Its size is fixed so
	// a single universal base fits every generated shade; the shade
	// self-centers when placed over the LED pocket.
	recessR := pocketR + 2
	recessDepth := 1.2
	recess, err := sdf.Cylinder3D(recessDepth*2, recessR, 0)
	if err != nil {
		return nil, err
	}
	recess = sdf.Transform3D(recess, sdf.Translate3d(v3.Vec{Z: p.BaseHeight - recessDepth}))
	body = sdf.Difference3D(body, recess)

	return body, nil
}

// cableSlot builds a small rectangular channel from the LED pocket floor
// through the outer wall so the USB cable can leave the base.
func cableSlot(p *design.Params, pocketR float64) (sdf.SDF3, error) {
	// Slot cross section: CableSlotWidth wide, 4 mm high.
	slot, err := sdf.Box3D(
		v3.Vec{X: p.BaseRadius - pocketR + 4, Y: p.CableSlotWidth, Z: 4},
		1.0,
	)
	if err != nil {
		return nil, err
	}
	pocketFloor := p.BaseHeight - p.PuckHeight - p.PocketDepthOffset
	slot = sdf.Transform3D(slot, sdf.Translate3d(v3.Vec{
		X: pocketR - 2 + (p.BaseRadius-pocketR+4)/2,
		Z: pocketFloor + 2,
	}))
	return slot, nil
}
