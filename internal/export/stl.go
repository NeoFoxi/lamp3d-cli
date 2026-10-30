// Package export writes the generated models to STL files and keeps
// output file names unique across runs.
package export

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
)

// WriteSTL renders an SDF3 to an STL file.
func WriteSTL(s sdf.SDF3, path string, res int) {
	render.ToSTL(s, path, render.NewMarchingCubesUniform(res))
}

// UniquePath returns path unchanged when it does not exist yet; otherwise a
// numeric suffix (-2, -3, ...) is appended so repeated runs (same seed and
// mode, invoked with -seed or --seed or picked in the CLI) never overwrite
// an existing model.
func UniquePath(path string) string {
	if _, err := os.Stat(path); err != nil {
		return path // does not exist: use it as is
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d%s", base, i, ext)
		if _, err := os.Stat(candidate); err != nil {
			log.Printf("%s already exists, writing %s instead", path, candidate)
			return candidate
		}
	}
}
